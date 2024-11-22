package tdx

import (
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
	"path/filepath"
	"time"
	"xorm.io/xorm"
)

func NewDB(dir, code string) (*DB, error) {
	filename := filepath.Join(dir, code+".db")
	db, err := sqlite.NewXorm(filename)
	if err != nil {
		return nil, err
	}
	if err = db.Sync2(
		new(Info),
		new(Trade),
		NewKlineTable("Minute"),
		NewKlineTable("5Minute"),
		NewKlineTable("15Minute"),
		NewKlineTable("30Minute"),
		NewKlineTable("Hour"),
		NewKlineTable("Day"),
		NewKlineTable("Week"),
		NewKlineTable("Month"),
		NewKlineTable("Quarter"),
		NewKlineTable("Year"),
	); err != nil {
		return nil, err
	}

	co, err := db.Count(new(Info))
	if err != nil {
		return nil, err
	}
	if co == 0 {
		if _, err = db.Insert(new(Info)); err != nil {
			return nil, err
		}
	}

	return &DB{dir: dir, code: code, db: db}, nil
}

type DB struct {
	dir  string
	code string
	db   *xorms.Engine
}

func (this *DB) Close() error {
	return this.db.Close()
}

// Update 更新数据
func (this *DB) Update(c *tdx.Client) error {
	for _, f := range []func(c *tdx.Client, code string) ([]*Kline, error){
		this.KlineMinute,
		this.Kline5Minute,
		this.Kline15Minute,
		this.Kline30Minute,
		this.KlineHour,
		//this.KlineDay,
		this.KlineWeek,
		this.KlineMonth,
		this.KlineQuarter,
		this.KlineYear,
	} {
		if _, err := f(c, this.code); err != nil {
			return err
		}
	}

	klines, err := this.KlineDay(c, this.code)
	if err != nil {
		return err
	}
	dates := []string(nil)
	for _, v := range klines {
		dates = append(dates, time.Unix(v.Unix, 0).Format("20060102"))
	}

	if _, err := this.Trade(c, this.code, dates); err != nil {
		return err
	}

	return nil
}

// GetInfo 获取信息
func (this *DB) GetInfo() (*Info, error) {
	info := new(Info)
	_, err := this.db.Get(info)
	return info, err
}

func (this *DB) KlineMinute(c *tdx.Client, code string) ([]*Kline, error) {
	return this.kline("Minute", code, c.GetKlineMinute)
}

func (this *DB) Kline5Minute(c *tdx.Client, code string) ([]*Kline, error) {
	return this.kline("5Minute", code, c.GetKline5Minute)
}

func (this *DB) Kline15Minute(c *tdx.Client, code string) ([]*Kline, error) {
	return this.kline("15Minute", code, c.GetKline15Minute)
}

func (this *DB) Kline30Minute(c *tdx.Client, code string) ([]*Kline, error) {
	return this.kline("30Minute", code, c.GetKline30Minute)
}

func (this *DB) KlineHour(c *tdx.Client, code string) ([]*Kline, error) {
	return this.kline("Hour", code, c.GetKlineHour)
}

func (this *DB) KlineDay(c *tdx.Client, code string) ([]*Kline, error) {
	return this.kline("Day", code, c.GetKlineDay)
}

func (this *DB) KlineWeek(c *tdx.Client, code string) ([]*Kline, error) {
	return this.kline("Week", code, c.GetKlineWeek)
}

func (this *DB) KlineMonth(c *tdx.Client, code string) ([]*Kline, error) {
	return this.kline("Month", code, c.GetKlineMonth)
}

func (this *DB) KlineQuarter(c *tdx.Client, code string) ([]*Kline, error) {
	return this.kline("Quarter", code, c.GetKlineQuarter)
}

func (this *DB) KlineYear(c *tdx.Client, code string) ([]*Kline, error) {
	return this.kline("Year", code, c.GetKlineYear)
}

func (this *DB) kline(suffix, code string, get func(code string, start, count uint16) (*protocol.KlineResp, error)) ([]*Kline, error) {

	//1. 连接数据库
	table := NewKlineTable(suffix)
	logs.Debug(table.TableName())

	//2. 查询数据库的数据
	cache := []*Kline(nil)
	err := this.db.Table(table).Find(&cache)
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
	err = this.db.SessionFunc(func(session *xorm.Session) error {
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

	cache = append(cache, list...)

	//5. 更新K线入库的时间,避免重复从服务器拉取,失败问题也不大
	_, err = this.db.Table("Info").Update(map[string]int64{table.tableName: time.Now().Unix()})
	logs.PrintErr(err)

	return cache, nil
}

func (this *DB) getInfo() (*Info, error) {
	info := new(Info)
	_, err := this.db.Get(info)
	return info, err
}

/*
Trade
@code 股票代码，例sh000001
@dates 股票的所有交易日期，格式20241106
*/
func (this *DB) Trade(c *tdx.Client, code string, dates []string) ([]*Trade, error) {
	if len(dates) == 0 {
		return nil, nil
	}

	//2. 查询最后的数据时间
	last := new(Trade)
	_, err := this.db.Desc("ID").Get(last)
	if err != nil {
		return nil, err
	}

	//3. 判断最后一条数据是否是15:00的,否则删除当天的数据
	full := last.Hour == 15 && last.Minute == 0
	if !full {
		if _, err := this.db.Where("Date=?", last.Date).Delete(new(Trade)); err != nil {
			return nil, err
		}
	}

	//4. 如果最后一条数据是今天的数据，直接返回
	if last.Date == dates[len(dates)-1] && full {
		list := []*Trade(nil)
		err = this.db.Where("Date=?", last.Date).Find(&list)
		return list, err
	}

	//5. 获取数据
	list := [][]*Trade(nil) //时间倒序的
	for i := len(dates) - 1; i > 0; i-- {
		date := dates[i]
		if date < last.Date || (!full && date == last.Date) {
			break
		}
		resp, err := c.GetHistoryMinuteTradeAll(date, code)
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
	err = this.db.SessionFunc(func(session *xorm.Session) error {
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
