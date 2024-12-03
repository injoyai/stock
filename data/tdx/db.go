package tdx

import (
	"errors"
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/goutil/times"
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/data/tdx/model"
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
		new(model.Update),
		new(model.Trade),
		model.NewKlineTable("Minute"),
		model.NewKlineTable("5Minute"),
		model.NewKlineTable("15Minute"),
		model.NewKlineTable("30Minute"),
		model.NewKlineTable("Hour"),
		model.NewKlineTable("Day"),
		model.NewKlineTable("Week"),
		model.NewKlineTable("Month"),
		model.NewKlineTable("Quarter"),
		model.NewKlineTable("Year"),
	); err != nil {
		return nil, err
	}

	co, err := db.Count(new(model.Update))
	if err != nil {
		return nil, err
	}
	if co == 0 {
		if _, err = db.Insert(new(model.Update)); err != nil {
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
func (this *DB) Update(pool *Pool, dates []string) error {
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
func (this *DB) GetInfo() (*model.Update, error) {
	info := new(model.Update)
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

func (this *DB) KlineMinute(pool *Pool) ([]*model.Kline, error) {
	return this.kline("Minute", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKlineMinute(code, start, count)
	}, times.IntegerMinute)
}

func (this *DB) Kline5Minute(pool *Pool) ([]*model.Kline, error) {
	return this.kline("5Minute", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKline5Minute(code, start, count)
	}, times.IntegerMinute)
}

func (this *DB) Kline15Minute(pool *Pool) ([]*model.Kline, error) {
	return this.kline("15Minute", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKline15Minute(code, start, count)
	}, times.IntegerMinute)
}

func (this *DB) Kline30Minute(pool *Pool) ([]*model.Kline, error) {
	return this.kline("30Minute", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKline30Minute(code, start, count)
	}, times.IntegerMinute)
}

func (this *DB) KlineHour(pool *Pool) ([]*model.Kline, error) {
	return this.kline("Hour", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKlineHour(code, start, count)
	}, times.IntegerHour)
}

func (this *DB) KlineDay(pool *Pool) ([]*model.Kline, error) {
	return this.kline("Day", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKlineDay(code, start, count)
	}, times.IntegerDay)
}

func (this *DB) KlineWeek(pool *Pool) ([]*model.Kline, error) {
	return this.kline("Week", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKlineWeek(code, start, count)
	}, times.IntegerWeek)
}

func (this *DB) KlineMonth(pool *Pool) ([]*model.Kline, error) {
	return this.kline("Month", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKlineMonth(code, start, count)
	}, times.IntegerMonth)
}

func (this *DB) KlineQuarter(pool *Pool) ([]*model.Kline, error) {
	return this.kline("Quarter", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKlineQuarter(code, start, count)
	}, times.IntegerQuarter)
}

func (this *DB) KlineYear(pool *Pool) ([]*model.Kline, error) {
	return this.kline("Year", func(code string, start, count uint16) (*protocol.KlineResp, error) {
		c, err := pool.Get()
		if err != nil {
			return nil, err
		}
		defer pool.Put(c)
		return c.GetKlineYear(code, start, count)
	}, times.IntegerYear)
}

func (this *DB) kline(suffix string, get func(code string, start, count uint16) (*protocol.KlineResp, error), dealTime func(t time.Time) time.Time) ([]*model.Kline, error) {

	//1. 连接数据库
	table := model.NewKlineTable(suffix)
	logs.Debug("更新:", table.TableName())

	//2. 查询数据库的数据
	cache := []*model.Kline(nil)
	err := this.db.Table(table).Find(&cache)
	if err != nil {
		logs.Err(err)
		return nil, err
	}

	last := new(model.Kline)
	if len(cache) > 0 {
		last = cache[len(cache)-1]   //获取最后一条数据,用于截止从服务器拉的数据
		cache = cache[:len(cache)-1] //去除最后一条数据,用拉取过来的数据更新掉
	}

	//3. 从服务器拉取数据
	list := []*model.Kline(nil)
	size := uint16(800)
	for start := uint16(0); ; start += size {
		resp, err := get(this.code, start, size)
		if err != nil {
			logs.Err(err)
			return nil, err
		}

		done := false
		ls := []*model.Kline(nil)
		for _, v := range resp.List {
			node := dealTime(v.Time)
			if last.Node <= node.Unix() {
				ls = append(ls, model.NewKline(this.code, v, node))
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
		logs.Err(err)
		return nil, err
	}

	cache = append(cache, list...)

	//5. 更新K线入库的时间,避免重复从服务器拉取,失败问题也不大
	_, err = this.db.Table("Update").Update(map[string]int64{table.TableName(): time.Now().Unix()})
	logs.PrintErr(err)

	return cache, nil
}

func (this *DB) getInfo() (*model.Update, error) {
	info := new(model.Update)
	_, err := this.db.Get(info)
	return info, err
}

/*
Trade
@code 股票代码，例sh000001
@dates 股票的所有交易日期，格式20241106
*/
func (this *DB) Trade(c *tdx.Client, code string, dates []string) ([]*model.Trade, error) {
	if len(dates) == 0 {
		return nil, nil
	}

	//2. 查询最后的数据时间
	last := new(model.Trade)
	_, err := this.db.Desc("ID").Get(last)
	if err != nil {
		return nil, err
	}

	//3. 判断最后一条数据是否是15:00的,否则删除当天的数据
	full := last.Hour == 15 && last.Minute == 0
	if !full {
		if _, err := this.db.Where("Date=?", last.Date).Delete(new(model.Trade)); err != nil {
			return nil, err
		}
	}

	//4. 如果最后一条数据是今天的数据，直接返回
	if last.Date == dates[len(dates)-1] && full {
		list := []*model.Trade(nil)
		err = this.db.Where("Date=?", last.Date).Find(&list)
		return list, err
	}

	//5. 获取数据
	list := [][]*model.Trade(nil) //时间倒序的
	for i := len(dates) - 1; i > 0; i-- {
		date := dates[i]
		if date < last.Date || (!full && date == last.Date) {
			break
		}
		resp, err := c.GetHistoryMinuteTradeAll(date, code)
		if err != nil {
			return nil, err
		}
		ls := []*model.Trade(nil)
		for _, v := range resp.List {
			ls = append(ls, model.NewTrade(code, date, v))
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
	Handler func(pool *Pool) ([]*model.Kline, error)
}
