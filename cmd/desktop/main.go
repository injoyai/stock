package main

import (
	"github.com/injoyai/base/chans"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/data"
	"github.com/injoyai/stock/data/tdx"
)

func main() {

	Run(
		func(s *Stray) {
			corn := task.New()
			mu := s.AddMenu().Disable().SetName("日志")
			_ = corn
			_ = mu

			//连接客户端
			c, err := tdx.Dial(tdx.Hosts, 10)
			logs.PanicErr(err)

			//每天下午16点进行数据更新
			corn.SetTask("update", "0 0 16 * * *", func() {
				isHoliday, err := data.TodayIsHoliday()
				logs.PanicErr(err)
				logs.PrintErr(update(c, mu, c.GetStockCodes(), isHoliday))
				mu.SetName(corn.GetTask("update").Next.String())
			})

			go func() {
				codes := c.GetStockCodes()
				//更新数据
				logs.PrintErr(update(c, mu, codes, false))
				mu.SetName(corn.GetTask("update").Next.String())
			}()

		},
		WithStartup(),
		WithSeparator(),
		WithExit(),
	)

}

func update(c *tdx.Client, mu *Menu, codes []string, isHoliday bool, retrys ...int) error {

	ch := chans.NewLimit(10)

	retry := conv.DefaultInt(3, retrys...)

	//1. 判断是否是节假日
	if isHoliday {
		return nil
	}

	mu.SetName("数据拉取中...")

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
			go func(f func(code string) ([]*tdx.Kline, error), code string) {
				ch.Add()
				defer ch.Done()
				g.Retry(func() error {
					_, err := f(code)
					logs.PrintErr(err)
					return err
				}, retry)
			}(f, code)

		}

	}

	return nil
}
