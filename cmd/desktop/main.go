package main

import (
	"github.com/getlantern/systray"
	"github.com/injoyai/goutil/g"
	"time"
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	//systray.SetIcon(IcoNotice)
	systray.SetTooltip("通达信数据拉取")
	nextUpdateTime := systray.AddMenuItem("", "")
	_ = nextUpdateTime

	mKlineMinute := systray.AddMenuItem("", "")
	mKlineMinute.Hide()
	mKline5Minute := systray.AddMenuItem("", "")
	mKline5Minute.Hide()
	mKline15Minute := systray.AddMenuItem("", "")
	mKline15Minute.Hide()
	mKline30Minute := systray.AddMenuItem("", "")
	mKline30Minute.Hide()
	mKlineHour := systray.AddMenuItem("", "")
	mKlineHour.Hide()
	mKlineDay := systray.AddMenuItem("", "")
	mKlineDay.Hide()
	mKlineWeek := systray.AddMenuItem("", "")
	mKlineWeek.Hide()
	mKlineMonth := systray.AddMenuItem("", "")
	mKlineMonth.Hide()
	mKlineQuarter := systray.AddMenuItem("", "")
	mKlineQuarter.Hide()
	mKlineYear := systray.AddMenuItem("", "")
	mKlineYear.Hide()

	update := false
	go func() {
		for range g.Interval(time.Second) {
			if update {

			}

		}
	}()

	exit := systray.AddMenuItem("退出", "退出")
	go func() {
		<-exit.ClickedCh
		systray.Quit()
	}()
}

func onExit() {

}
