package demo

import "github.com/injoyai/stock/strategy"

func init() {
	riseLimit := 0.05 //上涨阈值
	noticed := false
	strategy.All.Register(func(data *strategy.Data) {
		if data.Code != "sz000001" {
			return
		}
		if len(data.TodayKline) == 0 {
			return
		}
		max := float64(0)
		min := float64(0)
		open := data.TodayKline[0].Open
		for _, v := range data.TodayKline {
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
