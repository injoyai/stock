package tdx

import (
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/tdx"
)

func KlineDay(c *tdx.Client, code string) ([]string, error) {

	//1. 连接数据库
	filename := fmt.Sprintf("./database/kline/%s_day.db", code)
	db, err := sqlite.NewXorm(filename)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	//2. 查询数据库最后的数据
	data := new(StockKline)
	has, err := db.Desc("ID").Get(data)
	if err != nil {
		return nil, err
	}
	_ = has

	for {

	}

	//2. 查询最后的数据时间
	resp, err := c.GetKlineDayAll(code)
	if err != nil {
		return nil, err
	}

	dates := []string(nil)
	for _, v := range resp.List {
		dates = append(dates, v.Time.Format("20060102"))

	}

	return nil, nil
}

/*
Trade
@c 通达信的客户端
@code 股票代码，例sh000001
@dates 股票的所有交易日期，格式20241106
*/
func Trade(c *tdx.Client, code string, dates []string) error {

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
			resp, err := c.GetHistoryMinuteTradeAll(date, code)
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
