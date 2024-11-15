package tdx

import (
	"errors"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/goutil/times"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/data"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
	"github.com/robfig/cron/v3"
	"time"
	"xorm.io/xorm"
)

var Hosts = tdx.Hosts

func Dial(hosts []string, op ...client.Option) (*Client, error) {

	cli := &Client{}

	db, err := cli.OpenDB("update", new(Update))
	if err != nil {
		return nil, err
	}
	co, err := db.Count(new(Update))
	if err != nil {
		return nil, err
	}
	if co == 0 {
		_, err = db.Insert(new(Update))
		if err != nil {
			return nil, err
		}
	}

	cli.Client, err = tdx.DialWith(tdx.NewHostDial(hosts, 0), func(c *client.Client) {
		c.Logger.Debug()
		c.SetRedial(true)
		c.SetOption(op...)
	})
	if err != nil {
		return nil, err
	}

	cli.updateDB = db
	cli.Update = conv.NewExtend(cli)
	cli.Cron = cron.New(cron.WithSeconds())
	cli.Codes = make(map[string]*Code)

	//判断是否是节假日
	isHoliday, _ := data.TodayIsHoliday()
	codes, err := cli.Code(isHoliday)
	if err != nil {
		cli.Client.Close()
		return nil, err
	}

	for _, code := range codes {
		cli.Codes[code.Exchange+code.Code] = code
	}

	//每天4点更新代码信息,比如新增了股票,或者股票改了名字
	cli.Cron.AddFunc("0 0 4 * * *", func() {
		//1. 判断是否是节假日
		if isHoliday, _ := data.TodayIsHoliday(); isHoliday {
			return
		}

		//2. 更新代码信息
		codes, err = cli.Code(isHoliday)
		if err != nil {
			logs.Err(err)
			return
		}
		codeMap := make(map[string]*Code)
		for _, code := range codes {
			codeMap[code.Exchange+code.Code] = code
		}
		cli.Codes = codeMap
	})

	return cli, nil
}

/*
Client 客户端
*/
type Client struct {
	Client   *tdx.Client
	updateDB *xorms.Engine
	Update   conv.Extend
	Codes    map[string]*Code
	*cron.Cron
}

// GetCodes 获取股票代码
func (this *Client) GetCodes() []string {
	ls := make([]string, len(this.Codes))
	i := 0
	for k, _ := range this.Codes {
		ls[i] = k
		i++
	}
	return ls
}

// GetCodeName 获取股票中文名称
func (this *Client) GetCodeName(code string) string {
	if v, ok := this.Codes[code]; ok {
		return v.Name
	}
	return "未知"
}

// GetVar 实现接口
func (this *Client) GetVar(key string) *conv.Var {
	if this.updateDB == nil {
		return conv.Nil()
	}
	up := new(Update)
	if _, err := this.updateDB.Get(up); err != nil {
		return conv.Nil()
	}
	return up.GetVar(key)
}

// UpdateTime 更新时间,内部使用
func (this *Client) UpdateTime(key string) error {
	u := new(Update).Update(key)
	_, err := this.updateDB.Update(u)
	return err
}

// OpenDB 打开数据库,内部使用
func (this *Client) OpenDB(code string, entity any) (*xorms.Engine, error) {
	filename := "./database/" + code + ".db"
	db, err := sqlite.NewXorm(filename)
	if err != nil {
		return nil, err
	}
	if err := db.Sync(entity); err != nil {
		return nil, err
	}
	return db, nil
}

// Code 更新股票并返回结果
func (this *Client) Code(byDatabase bool) ([]*Code, error) {

	//1. 打开数据库
	db, err := this.OpenDB("code", new(Code))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	//2. 查询数据库所有股票
	list := []*Code(nil)
	if err := db.Find(&list); err != nil {
		return nil, err
	}

	//如果是从缓存读取,则返回结果
	if byDatabase {
		return list, nil
	}

	//判断今天是否更新过
	if times.IntegerDay(time.Now()).Unix() < this.Update.GetInt64("code") {
		return list, nil
	}

	mCode := make(map[string]*Code, len(list))
	for _, v := range list {
		mCode[v.Code] = v
	}

	//3. 从服务器获取所有股票代码
	insert := []*Code(nil)
	update := []*Code(nil)
	for _, exchange := range []protocol.Exchange{protocol.ExchangeSH, protocol.ExchangeSZ} {
		resp, err := this.Client.GetCodeAll(exchange)
		if err != nil {
			return nil, err
		}
		for _, v := range resp.List {
			if _, ok := mCode[v.Code]; ok {
				if mCode[v.Code].Name != v.Name {
					mCode[v.Code].Name = v.Name
					update = append(update, NewCode(exchange, v))
				}
			} else {
				code := NewCode(exchange, v)
				insert = append(insert, code)
				list = append(list, code)
			}
		}
	}

	//4. 插入或者更新数据库
	err = db.SessionFunc(func(session *xorm.Session) error {
		for _, v := range insert {
			if _, err := session.Insert(v); err != nil {
				return err
			}
		}
		for _, v := range update {
			if _, err := session.Where("Code=?", v.Code).Cols("Name").Update(v); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	//更新获取代码的时间点
	logs.PrintErr(this.UpdateTime("code"))

	return list, nil

}

// Quote 盘口信息
func (this *Client) Quote(code string) (*protocol.Quote, error) {
	resp, err := this.Client.GetQuote(code)
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 {
		return nil, errors.New("not found")
	}
	return resp[0], nil
}

// KlineReal 分钟实时获取
func (this *Client) KlineReal(code string, cache Klines) (Klines, error) {

	last := &Kline{Unix: times.IntegerDay(time.Now()).Unix()}
	if len(cache) > 0 {
		last = cache[len(cache)-1]   //获取最后的数据,用于截止获取数据
		cache = cache[len(cache)-1:] //删除最后一分钟的数据,最后一分钟实时统计的,用新数据更新
	}

	size := uint16(800)
	list := Klines(nil)
	for {
		resp, err := this.Client.GetKlineMinute(code, 0, size)
		if err != nil {
			return cache, err
		}

		done := false
		for _, v := range resp.List {
			//获取今天有效的分时图
			if last.Unix <= v.Time.Unix() {
				list = append(list, NewKline(code, v))
			} else {
				done = true
			}
		}

		if resp.Count < size || done {
			break
		}

	}

	cache = append(cache, list...)
	return cache, nil

}

func (this *Client) KlineMinute(code string) ([]*Kline, error) {
	return this.kline("Minute", code, this.Client.GetKlineMinute)
}

func (this *Client) Kline5Minute(code string) ([]*Kline, error) {
	return this.kline("5Minute", code, this.Client.GetKline5Minute)
}

func (this *Client) Kline15Minute(code string) ([]*Kline, error) {
	return this.kline("15Minute", code, this.Client.GetKline15Minute)
}

func (this *Client) Kline30Minute(code string) ([]*Kline, error) {
	return this.kline("30Minute", code, this.Client.GetKline30Minute)
}

func (this *Client) KlineHour(code string) ([]*Kline, error) {
	return this.kline("Hour", code, this.Client.GetKlineHour)
}

func (this *Client) KlineDay(code string) ([]*Kline, error) {
	return this.kline("Day", code, this.Client.GetKlineDay)
}

func (this *Client) KlineWeek(code string) ([]*Kline, error) {
	return this.kline("Week", code, this.Client.GetKlineWeek)
}

func (this *Client) KlineMonth(code string) ([]*Kline, error) {
	return this.kline("Month", code, this.Client.GetKlineMonth)
}

func (this *Client) KlineQuarter(code string) ([]*Kline, error) {
	return this.kline("Quarter", code, this.Client.GetKlineQuarter)
}

func (this *Client) KlineYear(code string) ([]*Kline, error) {
	return this.kline("Year", code, this.Client.GetKlineYear)
}

func (this *Client) kline(suffix, code string, get func(code string, start, count uint16) (*protocol.KlineResp, error)) ([]*Kline, error) {

	//1. 连接数据库
	table := NewKlineTable(suffix)
	db, err := this.OpenDB(code, table)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	//2. 查询数据库的数据
	cache := []*Kline(nil)
	err = db.Table(table).Find(&cache)
	if err != nil {
		return nil, err
	}

	last := new(Kline)
	if len(cache) > 0 {
		last = cache[len(cache)-1]   //获取最后一条数据,用于截止从服务器拉的数据
		cache = cache[:len(cache)-1] //去除最后一条数据,用拉取过来的数据更新掉
	}

	//3. 从服务器拉取数据
	list := []*Kline(nil)
	size := uint16(800)
	for start := uint16(0); ; start += size {
		resp, err := get(code, start, size)
		if err != nil {
			return nil, err
		}

		done := false
		ls := []*Kline(nil)
		for _, v := range resp.List {
			if last.Unix <= v.Time.Unix() {
				ls = append(ls, NewKline(code, v))
			} else {
				done = true
			}
		}
		list = append(ls, list...)
		if resp.Count < size || done {
			break
		}
	}

	//4. 将缺的数据入库
	err = db.SessionFunc(func(session *xorm.Session) error {
		for _, v := range list {
			if v.Unix == last.Unix {
				//更新数据库的最后一条数据
				if _, err := session.Table(table).Where("Unix=?", v.Unix).Update(v); err != nil {
					return err
				}
			} else {
				//插入新获取到的数据
				if _, err := session.Table(table).Insert(v); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	//5. 更新K线入库的时间,避免重复从服务器拉取,失败问题也不大
	logs.PrintErr(this.UpdateTime("Kline" + suffix))

	cache = append(cache, list...)

	return cache, nil
}

/*
Trade
@code 股票代码，例sh000001
@dates 股票的所有交易日期，格式20241106
*/
func (this *Client) Trade(code string, dates []string) ([]*Trade, error) {
	if len(dates) == 0 {
		return nil, nil
	}

	//1. 连接数据库
	db, err := this.OpenDB(code, new(Trade))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	//2. 查询最后的数据时间
	last := new(Trade)
	_, err = db.Desc("ID").Get(last)
	if err != nil {
		return nil, err
	}

	//3. 判断最后一条数据是否是15:00的,否则删除当天的数据
	full := last.Hour == 15 && last.Minute == 0
	if !full {
		if _, err := db.Where("Date=?", last.Date).Delete(new(Trade)); err != nil {
			return nil, err
		}
	}

	//4. 如果最后一条数据是今天的数据，直接返回
	if last.Date == dates[len(dates)-1] && full {
		list := []*Trade(nil)
		err = db.Where("Date=?", last.Date).Find(&list)
		return list, err
	}

	//5. 获取数据
	list := [][]*Trade(nil) //时间倒序的
	for i := len(dates) - 1; i > 0; i-- {
		date := dates[i]
		if date < last.Date || (!full && date == last.Date) {
			break
		}
		resp, err := this.Client.GetHistoryMinuteTradeAll(date, code)
		if err != nil {
			return nil, err
		}
		ls := []*Trade(nil)
		for _, v := range resp.List {
			ls = append(ls, NewTrade(code, date, v))
		}
		list = append(list, ls)
		if resp.Count == 0 {
			break
		}
	}

	//6. 插入到数据库
	err = db.SessionFunc(func(session *xorm.Session) error {
		for i := len(list) - 1; i >= 0; i-- {
			for _, v := range list[i] {
				if _, err := session.Insert(v); err != nil {
					return err
				}
			}
		}
		return nil
	})

	return list[0], nil
}
