package tdx

import (
	"errors"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/goutil/times"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
	"github.com/robfig/cron/v3"
	"os"
	"time"
	"xorm.io/xorm"
)

var (
	Hosts = tdx.Hosts
)

type Cli = tdx.Client

func Dial(hosts []string, cap int, op ...client.Option) (*Client, error) {

	cli := &Client{
		Database: "./database2/",
	}
	os.Mkdir(cli.Database, os.ModePerm)

	//增加一个自用客户端,用于获取股票代码信息
	c, err := tdx.DialWith(tdx.NewHostDial(hosts, 0), op...)
	if err != nil {
		return nil, err
	}
	db, err := cli.OpenDB("info", new(Info))
	if err != nil {
		return nil, err
	}
	cli.Workday, err = newWorkday(c, db)
	if err != nil {
		logs.Err(err)
		return nil, err
	}

	//新建连接池,
	cli.Pool, err = NewPool(hosts, cap, func(c *client.Client) {
		c.Logger.Debug()
		c.SetRedial(true)
		c.SetOption(op...)
	})
	if err != nil {
		return nil, err
	}

	update := func() error {
		logs.Debug("update")
		//1. 更新工作日数据
		err = cli.Workday.Update()
		logs.PrintErr(err)

		logs.Debug("IsWorkday")
		//2. 判断是否是节假日
		isWorkday := cli.Workday.TodayIs()
		if !isWorkday {
			return nil
		}
		logs.Debug("UpdateCode")
		//3. 更新代码信息
		return cli.UpdateCode(!isWorkday)
	}

	//启动更新一次
	if err := update(); err != nil {
		cli.Pool.Close()
		return nil, err
	}

	//每天4点更新代码信息,比如新增了股票,或者股票改了名字
	cron.New(cron.WithSeconds()).AddFunc("0 0 4 * * *", func() {
		logs.PrintErr(update())
	})

	return cli, nil
}

/*
Client 客户端
*/
type Client struct {
	Pool     *Pool
	Codes    map[string]*Code
	codeDB   *xorms.Engine //代码数据库实例
	Workday  *workday      //工作日
	Database string
}

func (this *Client) DB(code string) (*DB, error) {
	return NewDB(this.Database, code)
}

// UpdateCodes 更新股票
func (this *Client) UpdateCodes(codes []string, isHoliday bool, retrys ...int) error {
	retry := conv.DefaultInt(3, retrys...)

	//1. 判断是否是节假日
	if isHoliday {
		return nil
	}

	//2. 遍历全部股票
	for i := 0; i < len(codes); i++ {

		logs.Debug(codes[i])

		db, err := this.DB(codes[i])
		if err != nil {
			logs.Err(err)
			continue
		}

		return this.Pool.Retry(func(c *tdx.Client) error { return db.Update(c) }, retry)

	}

	return nil
}

// GetStockCodes 获取股票代码,不一定全
func (this *Client) GetStockCodes() []string {
	ls := []string(nil)
	for k, _ := range this.Codes {
		if len(k) == 8 {
			switch k[:2] {
			case "sz":
				if IsStock(protocol.ExchangeSZ, k[2:]) {
					ls = append(ls, k)
				}
			case "sh":
				if IsStock(protocol.ExchangeSH, k[2:]) {
					ls = append(ls, k)
				}
			}
		}
	}
	return ls
}

// GetCodes 获取所有代码
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

// OpenDB 打开数据库,内部使用
func (this *Client) OpenDB(code string, entity ...any) (*xorms.Engine, error) {
	filename := this.Database + code + ".db"
	db, err := sqlite.NewXorm(filename)
	if err != nil {
		return nil, err
	}
	if err := db.Sync(entity...); err != nil {
		return nil, err
	}
	return db, nil
}

func (this *Client) UpdateCode(byDatabase bool) error {
	codes, err := this.Code(byDatabase)
	if err != nil {
		return err
	}
	codeMap := make(map[string]*Code)
	for _, code := range codes {
		codeMap[code.Exchange+code.Code] = code
	}
	this.Codes = codeMap
	return nil
}

// Code 更新股票并返回结果
func (this *Client) Code(byDatabase bool) ([]*Code, error) {

	c := this.Pool.Get()
	defer this.Pool.Put(c)

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

	////判断今天是否更新过
	//if times.IntegerDay(time.Now()).Unix() < this.Update.GetInt64("code") {
	//	return list, nil
	//}

	mCode := make(map[string]*Code, len(list))
	for _, v := range list {
		mCode[v.Code] = v
	}

	//3. 从服务器获取所有股票代码
	insert := []*Code(nil)
	update := []*Code(nil)
	for _, exchange := range []protocol.Exchange{protocol.ExchangeSH, protocol.ExchangeSZ} {
		resp, err := c.GetCodeAll(exchange)
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

	return list, nil

}

// Quote 盘口信息
func (this *Client) Quote(code string) (*protocol.Quote, error) {
	c := this.Pool.Get()
	defer this.Pool.Put(c)
	resp, err := c.GetQuote(code)
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

	c := this.Pool.Get()
	defer this.Pool.Put(c)

	last := &Kline{Unix: times.IntegerDay(time.Now()).Unix()}
	if len(cache) > 0 {
		last = cache[len(cache)-1]   //获取最后的数据,用于截止获取数据
		cache = cache[len(cache)-1:] //删除最后一分钟的数据,最后一分钟实时统计的,用新数据更新
	}

	size := uint16(800)
	list := Klines(nil)
	for {
		resp, err := c.GetKlineMinute(code, 0, size)
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

///*
//Trade
//@code 股票代码，例sh000001
//@dates 股票的所有交易日期，格式20241106
//*/
//func (this *Client) Trade(code string, dates []string) ([]*Trade, error) {
//	if len(dates) == 0 {
//		return nil, nil
//	}
//
//	c := this.Pool.Get()
//	defer this.Pool.Put(c)
//
//	//1. 连接数据库
//	db, err := this.OpenDB(code, new(Trade))
//	if err != nil {
//		return nil, err
//	}
//	defer db.Close()
//
//	//2. 查询最后的数据时间
//	last := new(Trade)
//	_, err = db.Desc("ID").Get(last)
//	if err != nil {
//		return nil, err
//	}
//
//	//3. 判断最后一条数据是否是15:00的,否则删除当天的数据
//	full := last.Hour == 15 && last.Minute == 0
//	if !full {
//		if _, err := db.Where("Date=?", last.Date).Delete(new(Trade)); err != nil {
//			return nil, err
//		}
//	}
//
//	//4. 如果最后一条数据是今天的数据，直接返回
//	if last.Date == dates[len(dates)-1] && full {
//		list := []*Trade(nil)
//		err = db.Where("Date=?", last.Date).Find(&list)
//		return list, err
//	}
//
//	//5. 获取数据
//	list := [][]*Trade(nil) //时间倒序的
//	for i := len(dates) - 1; i > 0; i-- {
//		date := dates[i]
//		if date < last.Date || (!full && date == last.Date) {
//			break
//		}
//		resp, err := c.GetHistoryMinuteTradeAll(date, code)
//		if err != nil {
//			return nil, err
//		}
//		ls := []*Trade(nil)
//		for _, v := range resp.List {
//			ls = append(ls, NewTrade(code, date, v))
//		}
//		list = append(list, ls)
//		if resp.Count == 0 {
//			break
//		}
//	}
//
//	//6. 插入到数据库
//	err = db.SessionFunc(func(session *xorm.Session) error {
//		for i := len(list) - 1; i >= 0; i-- {
//			for _, v := range list[i] {
//				if _, err := session.Insert(v); err != nil {
//					return err
//				}
//			}
//		}
//		return nil
//	})
//
//	return list[0], nil
//}
//
//// IsHoliday 是否是节假日
//func (this *Client) IsHoliday(date string, countries ...string) (bool, error) {
//	t, err := time.Parse("20060102", date)
//	if err != nil {
//		return false, err
//	}
//
//	//周末
//	if t.Weekday() == 0 || t.Weekday() == 6 {
//		return true, nil
//	}
//
//	country := "中国"
//	if len(countries) > 0 {
//		country = countries[0]
//	}
//
//	_, month, day := t.Date()
//	switch {
//	case (country == "中国" && month == 1 && day == 1) ||
//		(country == "中国" && month == 10 && day >= 1 && day <= 7) ||
//		(country == "中国" && month == 5 && day >= 1 && day <= 3): //五一调休不一定从1-5号
//	}
//
//	list := []*Holiday(nil)
//	if err := this.codeDB.Find(&list); err != nil {
//		return false, err
//	}
//
//	m := make(map[string]struct{})
//	first := &Holiday{Date: times.Now().IntegerYear().Format("20060102")}
//	for i, v := range list {
//		if i == 0 {
//			first = v
//		}
//		m[v.Date] = struct{}{}
//	}
//
//	if date < first.Date {
//		return false, errors.New("没有之前的数据")
//	}
//
//	_, ok := m[country]
//
//	return ok, nil
//}
