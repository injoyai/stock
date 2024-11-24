package main

import (
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/data/tdx"
	"github.com/robfig/cron/v3"
	"path/filepath"
)

func init() {
	logs.SetShowColor(false)
	cfg.Init(cfg.WithFile(filepath.Join(oss.ExecDir(), "/config/config.yaml")))
}

func main() {

	Run(
		func(s *Stray) {
			go func() {
				task := cron.New(cron.WithSeconds())

				//连接客户端
				c, err := tdx.Dial(&tdx.Config{
					Hosts:    cfg.GetStrings("hosts"),
					Cap:      cfg.GetInt("cap", 10),
					Database: cfg.GetString("database", "./database/"),
					Workday:  cfg.GetString("workday", "workday"),
				})
				logs.PanicErr(err)

				//每天下午16点进行数据更新
				task.AddFunc("0 0 16 * * *", func() {
					if c.Workday.TodayIs() {
						notice.DefaultWindows.Publish(&notice.Message{
							Content: "开始更新数据...",
						})
						err = update(c, c.GetStockCodes())
						logs.PrintErr(err)
					}
				})

				//定时输出到csv
				task.AddFunc("0 0 18 * * *", func() {

				})

				codes := c.GetStockCodes()

				//更新数据
				logs.PrintErr(update(c, codes))
			}()
		},
		WithLabel("版本: v0.0.1"),
		WithStartup(),
		WithSeparator(),
		WithExit(),
	)

}

func update(c *tdx.Client, codes []string, retrys ...int) error {

	retry := conv.DefaultInt(3, retrys...)

	logs.Info("开始更新数据...")

	//2. 遍历全部股票
	for i := range codes {
		//3. 进行按股票进行每日更新,并尝试重试
		code := codes[i]
		go c.Pool.Retry(func(cli *tdx.Cli) error {
			return c.WithOpenDB(code, func(db *tdx.DB) error {
				for _, v := range db.AllKlineHandler() {
					_, err := v(cli)
					logs.PrintErr(err)
				}
				return nil
			})
		}, retry)
	}

	return nil
}
