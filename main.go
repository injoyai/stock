package main

import (
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/api"
	"github.com/injoyai/stock/data"
	"github.com/injoyai/stock/data/tdx"
)

func main() {

	//判断是否是节假日
	isHoliday, err := data.TodayIsHoliday()
	logs.PrintErr(err)

	//连接客户端
	c, err := tdx.Dial(tdx.Hosts)
	logs.PanicErr(err)

	//每天下午16点进行数据更新
	//common.Corn.SetTask("update", "0 0 16 * * *", func() {

	//1. 判断是否是节假日
	if isHoliday {
		return
	}

	codes := []*tdx.Code{{Code: "sz000001"}}

	//2. 遍历全部股票
	for _, code := range codes {
		//3. 进行按股票进行每日更新,并尝试重试
		fns := []func(code string) ([]*tdx.Kline, error){
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
			//c.Trade,
		}

		//fns = []func(code string) ([]*tdx.Kline, error){
		//	c.KlineMinute,
		//}

		for _, f := range fns {
			g.Retry(func() error {
				_, err = f(code.Code)
				logs.PrintErr(err)
				return err
			}, 3)
		}

	}

	//})

	//加载http服务
	logs.Err(api.Run(8080))
}
