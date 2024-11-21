package main

import (
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/logs"
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
			c.Database = "./dataabase2/"

			//每天下午16点进行数据更新
			corn.SetTask("update", "0 0 16 * * *", func() {
				isWorkday := c.Workday.TodayIs()
				logs.PanicErr(err)
				logs.PrintErr(update(c, mu, c.GetStockCodes(), !isWorkday))
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

	retry := conv.DefaultInt(3, retrys...)

	//1. 判断是否是节假日
	if isHoliday {
		return nil
	}

	mu.SetName("数据拉取中...")

	//2. 遍历全部股票
	for _, v := range codes {
		//3. 进行按股票进行每日更新,并尝试重试
		code := v

		db, err := c.DB(code)
		if err != nil {
			return err
		}

		for _, f := range []func(c *tdx.Cli, code string) ([]*tdx.Kline, error){
			db.KlineMinute,
			db.Kline5Minute,
			db.Kline15Minute,
			db.Kline30Minute,
			db.KlineHour,
			db.KlineDay,
			db.KlineWeek,
			db.KlineMonth,
			db.KlineQuarter,
			db.KlineYear,
		} {
			go c.Pool.Retry(func(c *tdx.Cli) error {
				_, err := f(c, code)
				logs.PrintErr(err)
				return err
			}, retry)

		}
		db.Close()

	}

	return nil
}
