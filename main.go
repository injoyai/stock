package main

import (
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/api"
	"github.com/injoyai/stock/common"
	"github.com/injoyai/stock/data"
	"github.com/injoyai/stock/data/tdx"
	"github.com/injoyai/stock/gui"
	"github.com/injoyai/stock/strategy"
)

func main() {

	//判断是否是节假日
	isHoliday, err := data.TodayIsHoliday()
	logs.PrintErr(err)

	//连接客户端
	c, err := tdx.Dial("124.71.187.122")
	logs.PanicErr(err)

	if true {

		ls, err := c.KlineReal("sz002244", tdx.Klines{})
		logs.PanicErr(err)
		//for _, v := range ls {
		//	logs.Info(v)
		//}

		gui.ShowChart(ls.Chart("曲线1"))

	}

	codes := []string(nil)
	if false {
		//启动的时候获取全部股票
		codes, err = c.Code(isHoliday)
		logs.PrintErr(err)
		logs.Info("更新全部股票结束...")
	}

	if false {
		//每天早上8点更新股票代码,或者是启动的时候
		common.Corn.SetTask("updateCode", "0 30 7 * * *", func() {
			codes, err = c.Code(isHoliday)
			logs.PrintErr(err)
		})
	}

	if false {
		//每天下午16点进行数据更新
		common.Corn.SetTask("update", "0 0 16 * * *", func() {

			//1. 判断是否是节假日
			if isHoliday {
				return
			}

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
	}

	if false {
		//关注的股票,或者全部股票
		codeReal := "sz000001"
		//今日分时k线图
		todayKline := tdx.Klines(nil)
		//今日分时成交
		todayTrace := []*tdx.MinuteTrade(nil)
		//每30秒更新分时数据,并实时计算
		common.Corn.SetTask("updateReal", "*/2 * * * * *", func() {

			//更新实时K线数据
			todayKline, err = c.KlineReal(codeReal, todayKline)
			logs.PrintErr(err)

			//更新实时分时成交
			_ = todayTrace

			//实时计算日策略
			strategy.All.Do(&strategy.Data{
				Code:         codeReal,
				TodayKline:   todayKline,
				TodayTrace:   todayTrace,
				HistoryKline: nil,
				HistoryTrace: nil,
			})

		})
	}

	//加载http服务
	logs.Err(api.Run(8080))
}
