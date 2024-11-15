package main

import (
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/common"
	"github.com/injoyai/stock/data"
	"github.com/injoyai/stock/data/tdx"
	"time"
)

func main() {

	//连接客户端
	c, err := tdx.Dial(tdx.Hosts)
	logs.PanicErr(err)

	//更新数据
	logs.PrintErr(update(c, c.GetCodes()))

	//每天下午16点进行数据更新
	common.Corn.SetTask("update", "0 0 16 * * *", func() {
		logs.PrintErr(update(c, c.GetCodes()))
	})

	//等待客户端退出
	<-c.Client.Done()
}

// 更新数据
func update(c *tdx.Client, codes []string) error {

	//1. 判断是否是节假日
	isHoliday, err := data.TodayIsHoliday()
	if err != nil {
		return err
	} else if isHoliday {
		return nil
	}

	//2. 遍历全部股票
	for _, code := range codes {
		//3. 进行按股票进行每日更新,并尝试重试
		for _, f := range []func(code string) ([]*tdx.Kline, error){
			c.KlineMinute,
			c.Kline5Minute,
			c.Kline15Minute,
			c.Kline30Minute,
			c.KlineHour,
			c.KlineDay,
			c.KlineWeek,
			c.KlineMonth,
			c.KlineQuarter,
			c.KlineYear,
		} {
			g.Retry(func() error {
				_, err := f(code)
				logs.PrintErr(err)
				return err
			}, 3)
		}

		//4. 获取日K线和所有日期
		dates := []string(nil)
		g.Retry(func() error {
			resp, err := c.KlineDay(code)
			if err != nil {
				logs.Err(err)
				return err
			}
			for _, v := range resp {
				dates = append(dates, time.Unix(v.Unix, 0).Format("20060102"))
			}
			return nil
		}, 3)

		//5. 获取分时成交
		g.Retry(func() error {
			_, err := c.Trade(code, dates)
			logs.PrintErr(err)
			return err
		}, 3)
	}

	return nil
}
