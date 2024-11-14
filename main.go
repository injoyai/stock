package main

import (
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/times"
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/api"
	"github.com/injoyai/stock/common"
	"github.com/injoyai/stock/data"
	"github.com/injoyai/stock/data/tdx"
	"github.com/injoyai/stock/strategy"
	"time"
)

func main() {

	//判断是否是节假日
	isHoliday, err := data.TodayIsHoliday()
	logs.PrintErr(err)

	//连接客户端
	c, err := tdx.Dial("124.71.187.122")
	logs.PanicErr(err)

	//每天下午16点进行数据更新
	common.Corn.SetTask("update", "0 0 16 * * *", func() {

		//1. 判断是否是节假日
		if isHoliday {
			return
		}

		codes := []*tdx.Code{{Code: "sz000001"}}

		//2. 遍历全部股票
		for _, code := range codes {
			//3. 进行按股票进行每日更新,并尝试重试
			g.Retry(func() error {
				_, err = c.KlineMinute(code.Code)
				logs.PrintErr(err)
				return err
			}, 3)
		}

	})

	if false {
		//关注的股票,或者全部股票
		codeReal := "sz000001"
		//今日分时k线图
		todayKline := tdx.Klines(nil)
		//今日分时成交
		todayTrace := []*tdx.MinuteTrade(nil)
		//每秒更新分时数据,并实时计算
		common.Corn.SetTask("updateReal", "* * 9-12,13-15 * * *", func() {

			//1. 判断是否是节假日
			if isHoliday {
				return
			}

			//2. 判断是否在交易时间内
			now := time.Now()
			start1 := time.Hour*9 + time.Minute*30
			end1 := time.Hour*11 + time.Minute*30
			start2 := time.Hour * 13
			end2 := time.Hour * 15
			sub := now.Sub(times.IntegerDay(now))
			if sub < start1 || sub > end2 || (sub > end1 && sub < start2) {
				return
			}

			//更新实时K线数据
			todayKline, err = c.KlineReal(codeReal, todayKline)
			logs.PrintErr(err)

			//实时计算日策略
			strategy.All.Do(&strategy.Data{
				Code:       codeReal,
				TodayKline: todayKline,
				TodayTrace: todayTrace,
			})

		})
	}

	//加载http服务
	logs.Err(api.Run(8080))
}
