package tdx

import (
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/goutil/times"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
	"time"
	"xorm.io/xorm"
)

func Dial(addr string, op ...client.Option) (*Client, error) {
	c, err := tdx.Dial(addr, func(c *client.Client) {
		c.Logger.Debug()
		c.SetRedial(true)
		c.SetOption(op...)
	})
	if err != nil {
		return nil, err
	}
	return &Client{c}, nil
}

type Client struct {
	*tdx.Client
}

func (this *Client) Code(byDatabase bool) ([]string, error) {

	//1. 打开数据库
	db, err := sqlite.NewXorm("./database/code.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	//2. 查询数据库所有股票
	list := []*StockCode(nil)
	if err := db.Find(&list); err != nil {
		return nil, err
	}
	codes := make([]string, len(list), len(list)+10)
	mCode := make(map[string]string, len(list))
	for i, v := range list {
		mCode[v.Code] = v.Name
		codes[i] = v.Code
	}

	//如果是从缓存读取,则返回结果
	if byDatabase {
		return codes, nil
	}

	//3. 从服务器获取所有股票代码
	insert := []*StockCode(nil)
	update := []*StockCode(nil)
	for _, exchange := range []protocol.Exchange{protocol.ExchangeSH, protocol.ExchangeSZ} {
		resp, err := this.Client.GetCodeAll(exchange)
		if err != nil {
			return nil, err
		}
		for _, v := range resp.List {
			if _, ok := mCode[v.Code]; ok {
				if mCode[v.Code] != v.Name {
					update = append(update, &StockCode{
						Name:     v.Name,
						Code:     v.Code,
						Exchange: v.Code[:2],
					})
				} else {
					insert = append(insert, &StockCode{
						Name:     v.Name,
						Code:     v.Code,
						Exchange: v.Code[:2],
					})
					codes = append(codes, v.Code)
				}
			}
		}
	}

	//4. 插入或者更新数据库
	return codes, db.SessionFunc(func(session *xorm.Session) error {
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

}

func (this *Client) GetKlineReal(code string, cache []*StockKline) ([]*StockKline, error) {

	last := &StockKline{Unix: times.IntegerDay(time.Now()).Unix()}
	if len(cache) > 0 {
		last = cache[len(cache)-1]   //获取最后的数据,用于截止获取数据
		cache = cache[len(cache)-1:] //删除最后一分钟的数据,用新数据更新
	}

	list := []*StockKline(nil)
	for {
		resp, err := this.Client.GetKlineMinute(code, 0, 800)
		if err != nil {
			return cache, err
		}

		done := false
		for _, v := range resp.List {
			//获取今天有效的分时图
			if last.Unix > v.Time.Unix() {
				done = true
				break
			}
			list = append(list, NewStockKline(code, v))
		}

		if done {
			break
		}

	}

	cache = append(cache, list...)
	return cache, nil

}

func (this *Client) KlineMinute(code string) error {
	return this.kline("minute", code, this.Client.GetKlineMinute)
}

func (this *Client) Kline5Minute(code string) error {
	return this.kline("5minute", code, this.Client.GetKline5Minute)
}

func (this *Client) Kline15Minute(code string) error {
	return this.kline("15minute", code, this.Client.GetKline15Minute)
}

func (this *Client) Kline30Minute(code string) error {
	return this.kline("30minute", code, this.Client.GetKline30Minute)
}

func (this *Client) KlineHour(code string) error {
	return this.kline("hour", code, this.Client.GetKlineHour)
}

func (this *Client) KlineDay(code string) error {
	return this.kline("day", code, this.Client.GetKlineDay)
}

func (this *Client) KlineWeek(code string) error {
	return this.kline("week", code, this.Client.GetKlineWeek)
}

func (this *Client) KlineMonth(code string) error {
	return this.kline("month", code, this.Client.GetStockKlineMonth) //todo 库的名字没改掉
}

func (this *Client) KlineQuarter(code string) error {
	return this.kline("quarter", code, this.Client.GetKlineQuarter)
}

func (this *Client) KlineYear(code string) error {
	return this.kline("year", code, this.Client.GetKlineYear)
}

func (this *Client) kline(suffix, code string, get func(code string, start, count uint16) (*protocol.KlineResp, error)) error {

	//1. 连接数据库
	filename := fmt.Sprintf("./database/kline/%s_%s.db", code, suffix)
	db, err := sqlite.NewXorm(filename)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Sync(new(StockKline)); err != nil {
		return err
	}

	//2. 查询数据库最后的数据
	last := new(StockKline)
	_, err = db.Desc("ID").Get(last)
	if err != nil {
		return err
	}

	lastTime := time.Unix(last.Unix, 0)
	list := []*StockKline(nil)

	size := uint16(800)
	for start := uint16(0); ; start += size {
		resp, err := get(code, start, size)
		if err != nil {
			return err
		}

		done := false
		ls := []*StockKline(nil)
		for _, v := range resp.List {
			if lastTime.Unix() < v.Time.Unix() {
				ls = append(ls, &StockKline{
					Exchange: code[:2],
					Code:     code[2:],
					Unix:     v.Time.Unix(),
					Year:     v.Time.Year(),
					Month:    int(v.Time.Month()),
					Day:      v.Time.Day(),
					Hour:     v.Time.Hour(),
					Minute:   v.Time.Minute(),
					Open:     v.Open.Float64(),
					High:     v.High.Float64(),
					Low:      v.Low.Float64(),
					Close:    v.Close.Float64(),
					Volume:   int64(v.Volume),
					Amount:   int64(v.Amount),
				})
			} else {
				done = true
			}
		}
		list = append(ls, list...)
		if resp.Count < size || done {
			break
		}
	}

	return db.SessionFunc(func(session *xorm.Session) error {
		for _, v := range list {
			if _, err := session.Insert(v); err != nil {
				return err
			}
		}
		return nil
	})

}

/*
KlineDay2 日k线
比KlineDay多返回个全部日期,用于判断有效地数据时间
*/
func (this *Client) KlineDay2(code string) ([]string, error) {

	//1. 连接数据库
	filename := fmt.Sprintf("./database/kline/%s_day.db", code)
	db, err := sqlite.NewXorm(filename)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if err := db.Sync(new(StockKline)); err != nil {
		return nil, err
	}

	//2. 查询数据库最后的数据
	last := new(StockKline)
	has, err := db.Desc("ID").Get(last)
	if err != nil {
		return nil, err
	}
	_ = has

	//2. 查询最后的数据时间
	resp, err := this.Client.GetKlineDayAll(code)
	if err != nil {
		return nil, err
	}

	list := []*StockKline(nil)
	dates := []string(nil)
	for _, v := range resp.List {
		dates = append(dates, v.Time.Format("20060102"))

		if last.Unix < v.Time.Unix() {
			list = append(list, &StockKline{
				Exchange: code[:2],
				Code:     code[2:],
				Unix:     v.Time.Unix(),
				Year:     v.Time.Year(),
				Month:    int(v.Time.Month()),
				Day:      v.Time.Day(),
				Hour:     v.Time.Hour(),
				Minute:   v.Time.Minute(),
				Open:     v.Open.Float64(),
				High:     v.High.Float64(),
				Low:      v.Low.Float64(),
				Close:    v.Close.Float64(),
				Volume:   int64(v.Volume),
				Amount:   int64(v.Amount),
			})
		}

	}

	err = db.SessionFunc(func(session *xorm.Session) error {
		for _, v := range list {
			if _, err := session.Insert(v); err != nil {
				return err
			}
		}
		return nil
	})

	return dates, err
}

/*
Trade
@c 通达信的客户端
@code 股票代码，例sh000001
@dates 股票的所有交易日期，格式20241106
*/
func (this *Client) Trade(code string, dates []string) error {

	//1. 连接数据库
	filename := fmt.Sprintf("./database/trade/%s_minute.db", code)
	db, err := sqlite.NewXorm(filename)
	if err != nil {
		return err
	}
	defer db.Close()

	//2. 查询最后的数据时间
	last := new(StockMinuteTrade)
	_, err = db.Desc("ID").Get(last)
	if err != nil {
		return err
	}

	//3. 遍历所有日期，判断是否有缺的数据
	for _, date := range dates {
		if last.Date < date {

			//4. 获取数据并插入
			resp, err := this.Client.GetHistoryMinuteTradeAll(date, code)
			if err != nil {
				return err
			}
			list := []*StockMinuteTrade(nil)
			for _, v := range resp.List {
				list = append(list, &StockMinuteTrade{
					Exchange: code[:2],
					Code:     code[2:],
					Date:     date,
					Year:     conv.Int(date[:4]),
					Month:    conv.Int(date[4:6]),
					Day:      conv.Int(date[6:8]),
					Hour:     conv.Int(v.Time[:2]),
					Minute:   conv.Int(v.Time[3:5]),
					Second:   0,
					Price:    v.Price.Float64(),
					Volume:   v.Volume,
					Number:   0,
					Status:   v.Status,
				})
			}
		}

	}

	return nil
}
