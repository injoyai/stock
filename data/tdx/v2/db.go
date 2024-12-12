package tdx

import (
	"errors"
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/goutil/times"
	"github.com/injoyai/stock/data/tdx/model"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
	"path/filepath"
	"time"
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

	return &DB{dir: dir, Code: code, db: db}, nil
}

type DB struct {
	dir  string
	Code string
	db   *xorms.Engine
}

func (this *DB) Close() error {
	return this.db.Close()
}

// Quote 盘口信息
func (this *DB) Quote(c *tdx.Client) (*protocol.Quote, error) {
	resp, err := c.GetQuote(this.Code)
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 {
		return nil, errors.New("not found")
	}
	return resp[0], nil
}

type Handler struct {
	Name    string
	Table   string
	Handler func(c *tdx.Client, code string, start, count uint16) (*protocol.KlineResp, error)
	Node    func(t time.Time) time.Time
}

var Handlers = []*Handler{
	{
		Name:  "1分K线",
		Table: "KlineMinute",
		Handler: func(c *tdx.Client, code string, start, count uint16) (*protocol.KlineResp, error) {
			return c.GetKlineMinute(code, start, count)
		},
		Node: times.IntegerMinute,
	},
	{
		Name:  "5分K线",
		Table: "Kline5Minute",
		Handler: func(c *tdx.Client, code string, start, count uint16) (*protocol.KlineResp, error) {
			return c.GetKline5Minute(code, start, count)
		},
		Node: times.IntegerMinute,
	},
	{
		Name:  "15分K线",
		Table: "Kline15Minute",
		Handler: func(c *tdx.Client, code string, start, count uint16) (*protocol.KlineResp, error) {
			return c.GetKline15Minute(code, start, count)
		},
		Node: times.IntegerMinute,
	},
	{
		Name:  "30分K线",
		Table: "Kline30Minute",
		Handler: func(c *tdx.Client, code string, start, count uint16) (*protocol.KlineResp, error) {
			return c.GetKline30Minute(code, start, count)
		},
		Node: times.IntegerMinute,
	},
	{
		Name:  "时K线",
		Table: "KlineHour",
		Handler: func(c *tdx.Client, code string, start, count uint16) (*protocol.KlineResp, error) {
			return c.GetKlineHour(code, start, count)
		},
		Node: times.IntegerHour,
	},
	{
		Name:  "日K线",
		Table: "KlineDay",
		Handler: func(c *tdx.Client, code string, start, count uint16) (*protocol.KlineResp, error) {
			return c.GetKlineDay(code, start, count)
		},
		Node: times.IntegerDay,
	},
	{
		Name:  "周K线",
		Table: "KlineWeek",
		Handler: func(c *tdx.Client, code string, start, count uint16) (*protocol.KlineResp, error) {
			return c.GetKlineWeek(code, start, count)
		},
		Node: times.IntegerWeek,
	},
	{
		Name:  "月K线",
		Table: "KlineMonth",
		Handler: func(c *tdx.Client, code string, start, count uint16) (*protocol.KlineResp, error) {
			return c.GetKlineMonth(code, start, count)
		},
		Node: times.IntegerMonth,
	},
	{
		Name:  "季K线",
		Table: "KlineQuarter",
		Handler: func(c *tdx.Client, code string, start, count uint16) (*protocol.KlineResp, error) {
			return c.GetKlineQuarter(code, start, count)
		},
		Node: times.IntegerQuarter,
	},
	{
		Name:  "年K线",
		Table: "KlineYear",
		Handler: func(c *tdx.Client, code string, start, count uint16) (*protocol.KlineResp, error) {
			return c.GetKlineYear(code, start, count)
		},
		Node: times.IntegerYear,
	},
}

var HandlerMap = func() map[string]*Handler {
	m := make(map[string]*Handler)
	for _, v := range Handlers {
		m[v.Table] = v
	}
	return m
}()

func (this *DB) Tables() []string {
	hs := Handlers
	ls := make([]string, len(hs))
	for i := range hs {
		ls[i] = hs[i].Table
	}
	return ls
}

func (this *DB) GetCache() (*DBData, error) {

	data := make(map[string][]*model.Kline)

	for _, table := range this.Tables() {
		cache := []*model.Kline(nil)
		err := this.db.Table(table).Find(&cache)
		if err != nil {
			return nil, err
		}
		data[table] = cache
	}

	return &DBData{
		DB:   this,
		Data: data,
	}, nil

}

///*
//Trade
//@code 股票代码，例sh000001
//@dates 股票的所有交易日期，格式20241106
//*/
//func (this *DB) Trade(c *tdx.Client, code string, dates []string) ([]*model.Trade, error) {
//	if len(dates) == 0 {
//		return nil, nil
//	}
//
//	//2. 查询最后的数据时间
//	last := new(model.Trade)
//	_, err := this.db.Desc("ID").Get(last)
//	if err != nil {
//		return nil, err
//	}
//
//	//3. 判断最后一条数据是否是15:00的,否则删除当天的数据
//	full := last.Hour == 15 && last.Minute == 0
//	if !full {
//		if _, err := this.db.Where("Date=?", last.Date).Delete(new(model.Trade)); err != nil {
//			return nil, err
//		}
//	}
//
//	//4. 如果最后一条数据是今天的数据，直接返回
//	if last.Date == dates[len(dates)-1] && full {
//		list := []*model.Trade(nil)
//		err = this.db.Where("Date=?", last.Date).Find(&list)
//		return list, err
//	}
//
//	//5. 获取数据
//	list := [][]*model.Trade(nil) //时间倒序的
//	for i := len(dates) - 1; i > 0; i-- {
//		date := dates[i]
//		if date < last.Date || (!full && date == last.Date) {
//			break
//		}
//		resp, err := c.GetHistoryMinuteTradeAll(date, code)
//		if err != nil {
//			return nil, err
//		}
//		ls := []*model.Trade(nil)
//		for _, v := range resp.List {
//			ls = append(ls, model.NewTrade(code, date, v))
//		}
//		list = append(list, ls)
//		if resp.Count == 0 {
//			break
//		}
//	}
//
//	//6. 插入到数据库
//	err = this.db.SessionFunc(func(session *xorm.Session) error {
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
