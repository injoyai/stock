package main

import (
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/api"
	"github.com/injoyai/stock/common"
	"github.com/injoyai/stock/data"
	"github.com/injoyai/stock/data/tdx"
	"time"
)

func main() {

	date := time.Now().Format("20060102")
	isHoliday, err := data.Holiday.Is(date)
	logs.PrintErr(err)

	//连接客户端
	c, err := tdx.Dial("124.71.187.122")
	logs.PanicErr(err)

	//启动的时候获取全部股票
	codes, err := c.Code(isHoliday)
	logs.PrintErr(err)

	//每天早上8点更新股票代码,或者是启动的时候
	common.Corn.SetTask("updateCode", "0 30 8 * * *", func() {
		codes, err = c.Code(isHoliday)
		logs.PrintErr(err)
	})

	//每天下午16点进行数据更新
	common.Corn.SetTask("update", "0 0 16 * * *", func() {

		//1. 判断是否是节假日

		//2. 遍历全部股票
		for _, code := range codes {
			//3. 进行按股票进行每日更新,并尝试重试
			g.Retry(func() error {
				err = c.KlineMinute(code)
				logs.PrintErr(err)
				return err
			}, 3)
		}

	})

	//今日分时k线图
	todayKline := []*tdx.StockKline(nil)
	//今日分时成交
	todayTrace := []*tdx.StockMinuteTrade(nil)
	//每秒更新实时数据,并实时计算
	common.Corn.SetTask("updateReal", "* * * * * *", func() {

		_ = todayKline
		_ = todayTrace

		//实时计算日策略

	})

	//加载http服务
	logs.Err(api.Run())
}
