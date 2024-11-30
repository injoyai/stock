package tdx

import (
	"errors"
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/logs"
	v1 "github.com/injoyai/stock/data/tdx"
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
		new(v1.Update),
		new(v1.Trade),
		v1.NewKlineTable("Minute"),
		v1.NewKlineTable("5Minute"),
		v1.NewKlineTable("15Minute"),
		v1.NewKlineTable("30Minute"),
		v1.NewKlineTable("Hour"),
		v1.NewKlineTable("Day"),
		v1.NewKlineTable("Week"),
		v1.NewKlineTable("Month"),
		v1.NewKlineTable("Quarter"),
		v1.NewKlineTable("Year"),
	); err != nil {
		return nil, err
	}

	co, err := db.Count(new(v1.Update))
	if err != nil {
		return nil, err
	}
	if co == 0 {
		if _, err = db.Insert(new(v1.Update)); err != nil {
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

func (this *DB) AllKlineHandler() []*Handler {
	return []*Handler{
		{"1分K线", this.KlineMinute},
		{"15分K线", this.Kline15Minute},
		{"30分K线", this.Kline30Minute},
		{"时K线", this.KlineHour},
		{"日K线", this.KlineDay},
		{"周K线", this.KlineWeek},
		{"月K线", this.KlineMonth},
		{"季K线", this.KlineQuarter},
		{"年K线", this.KlineYear},
	}
}

func (this *DB) Close() error {
	return this.db.Close()
}

// Update 更新数据
func (this *DB) Update(pool *v1.Pool, dates []string) error {
	for _, f := range this.AllKlineHandler() {
		if _, err := f.Handler(pool); err != nil {
			return err
		}
	}
	return pool.Do(func(c *tdx.Client) error {
		_, err := this.Trade(c, this.code, dates)
		return err

	})
}

// GetInfo 获取信息
func (this *DB) GetInfo() (*v1.Update, error) {
	info := new(v1.Update)
	_, err := this.db.Get(info)
	return info, err
}

// Quote 盘口信息
func (this *DB) Quote(c *tdx.Client) (*protocol.Quote, error) {
	resp, err := c.GetQuote(this.code)
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 {
		return nil, errors.New("not found")
	}
	return resp[0], nil
}

func (this *DB) KlineMinute(pool *v1.Pool) ([]*v1.Kline, error) {
	return this.kline("Minute", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKlineMinute(code, start, count)
	})
}

func (this *DB) Kline5Minute(pool *v1.Pool) ([]*v1.Kline, error) {
	return this.kline("5Minute", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKline5Minute(code, start, count)
	})
}

func (this *DB) Kline15Minute(pool *v1.Pool) ([]*v1.Kline, error) {
	return this.kline("15Minute", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKline15Minute(code, start, count)
	})
}

func (this *DB) Kline30Minute(pool *v1.Pool) ([]*v1.Kline, error) {
	return this.kline("30Minute", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKline30Minute(code, start, count)
	})
}

func (this *DB) KlineHour(pool *v1.Pool) ([]*v1.Kline, error) {
	return this.kline("Hour", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKlineHour(code, start, count)
	})
}

func (this *DB) KlineDay(pool *v1.Pool) ([]*v1.Kline, error) {
	return this.kline("Day", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKlineDay(code, start, count)
	})
}

func (this *DB) KlineWeek(pool *v1.Pool) ([]*v1.Kline, error) {
	return this.kline("Week", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKlineWeek(code, start, count)
	})
}

func (this *DB) KlineMonth(pool *v1.Pool) ([]*v1.Kline, error) {
	return this.kline("Month", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKlineMonth(code, start, count)
	})
}

func (this *DB) KlineQuarter(pool *v1.Pool) ([]*v1.Kline, error) {
	return this.kline("Quarter", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKlineQuarter(code, start, count)
	})
}

func (this *DB) KlineYear(pool *v1.Pool) ([]*v1.Kline, error) {
	return this.kline("Year", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKlineYear(code, start, count)
	})
}

func (this *DB) kline(suffix string, get func(code string, start, count uint16) (*protocol.KlineResp, error)) ([]*v1.Kline, error) {

	//1. 连接数据库
	table := v1.NewKlineTable(suffix)
	logs.Debug("更新:", table.TableName())

	//2. 查询数据库的数据
	cache := []*v1.Kline(nil)
	err := this.db.Table(table).Find(&cache)
	if err != nil {
		return nil, err
	}

	last := new(v1.Kline)
	if len(cache) > 0 {
		last = cache[len(cache)-1]   //获取最后一条数据,用于截止从服务器拉的数据
		cache = cache[:len(cache)-1] //去除最后一条数据,用拉取过来的数据更新掉
	}

	//3. 从服务器拉取数据
	list := []*v1.Kline(nil)
	size := uint16(800)
	for start := uint16(0); ; start += size {
		resp, err := get(this.code, start, size)
		if err != nil {
			return nil, err
		}

		done := false
		ls := []*v1.Kline(nil)
		for _, v := range resp.List {
			if last.Unix <= v.Time.Unix() {
				ls = append(ls, v1.NewKline(this.code, v))
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
	_, err = this.db.Table("Update").Update(map[string]int64{table.TableName(): time.Now().Unix()})
	logs.PrintErr(err)

	return cache, nil
}

func (this *DB) getInfo() (*v1.Update, error) {
	info := new(v1.Update)
	_, err := this.db.Get(info)
	return info, err
}

/*
Trade
@code 股票代码，例sh000001
@dates 股票的所有交易日期，格式20241106
*/
func (this *DB) Trade(c *tdx.Client, code string, dates []string) ([]*v1.Trade, error) {
	if len(dates) == 0 {
		return nil, nil
	}

	//2. 查询最后的数据时间
	last := new(v1.Trade)
	_, err := this.db.Desc("ID").Get(last)
	if err != nil {
		return nil, err
	}

	//3. 判断最后一条数据是否是15:00的,否则删除当天的数据
	full := last.Hour == 15 && last.Minute == 0
	if !full {
		if _, err := this.db.Where("Date=?", last.Date).Delete(new(v1.Trade)); err != nil {
			return nil, err
		}
	}

	//4. 如果最后一条数据是今天的数据，直接返回
	if last.Date == dates[len(dates)-1] && full {
		list := []*v1.Trade(nil)
		err = this.db.Where("Date=?", last.Date).Find(&list)
		return list, err
	}

	//5. 获取数据
	list := [][]*v1.Trade(nil) //时间倒序的
	for i := len(dates) - 1; i > 0; i-- {
		date := dates[i]
		if date < last.Date || (!full && date == last.Date) {
			break
		}
		resp, err := c.GetHistoryMinuteTradeAll(date, code)
		if err != nil {
			return nil, err
		}
		ls := []*v1.Trade(nil)
		for _, v := range resp.List {
			ls = append(ls, v1.NewTrade(code, date, v))
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

type Handler struct {
	Name    string
	Handler func(pool *v1.Pool) ([]*v1.Kline, error)
}
