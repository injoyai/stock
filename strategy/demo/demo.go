package demo

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/strategy"
)

func init() {
	riseLimit := 0.05 //上涨阈值
	noticed := false
	code := "sz000001"
	strategy.All.Register(func(i strategy.Interface) {

		TodayKline, err := i.GetKlineMinute(code)
		if err != nil {
			logs.Err(err)
			return
		}

		if len(TodayKline) == 0 {
			return
		}

		max := float64(0)
		min := float64(0)
		open := TodayKline[0].Open
		for _, v := range TodayKline {
			if v.High > max {
				max = v.High
			}
			if v.Low < min {
				min = v.Low
			}
		}
		if open > 0 && (max-min)/open > riseLimit {
			if noticed {
				//todo 发送通知
			}
		} else {
			noticed = false
		}
	})
}
