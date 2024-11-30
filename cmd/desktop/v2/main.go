package main

import (
	"github.com/injoyai/base/chans"
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/tray"
	"github.com/injoyai/goutil/other/csv"
	"github.com/injoyai/logs"
	v1 "github.com/injoyai/stock/data/tdx"
	"github.com/injoyai/stock/data/tdx/v2"
	"github.com/robfig/cron/v3"
	"path/filepath"
	"time"
)

func init() {
	logs.SetShowColor(false)
	cfg.Init(cfg.WithFile(filepath.Join(oss.ExecDir(), "/config/config.yaml")))
}

func main() {

	conf := &tdx.Config{
		Hosts:    cfg.GetStrings("hosts"),
		Number:   cfg.GetInt("number", 20),
		Database: cfg.GetString("database"),
	}

	tray.Run(
		func(s *tray.Stray) {
			go func() {
				task := cron.New(cron.WithSeconds())

				//连接客户端
				c, err := tdx.Dial(conf)
				logs.PanicErr(err)

				//每天下午16点进行数据更新
				task.AddFunc("0 0 16 * * *", func() {
					if c.Workday.TodayIs() {
						notice.DefaultWindows.Publish(&notice.Message{
							Content: "开始更新数据...",
						})
						err = update(c, c.Code.GetStocks(), conf.Number)
						logs.PrintErr(err)
					}
				})

				//更新数据
				codes := c.Code.GetStocks()
				logs.PrintErr(update(c, codes, conf.Number))
			}()
		},
		tray.WithLabel("版本: v0.2.0"),
		tray.WithStartup(),
		tray.WithSeparator(),
		tray.WithExit(),
		tray.WithHint("定时拉取股票信息"),
	)

}

func update(c *tdx.Client, codes []string, limit int, retries ...int) error {

	retry := conv.DefaultInt(3, retries...)

	logs.Info("开始更新数据...")

	ch := chans.NewWaitLimit(uint(limit))

	//2. 遍历全部股票
	for i := range codes {
		//3. 进行按股票进行每日更新,并尝试重试
		ch.Add()
		go func(code string) {
			defer ch.Done()
			c.WithOpenDB(code, func(db *tdx.DB) error {
				for _, v := range db.AllKlineHandler() {
					err := g.Retry(func() error {
						kline, err := v.Handler(c.Pool)
						if err != nil {
							return err
						}
						toCsv(c, filepath.Join(v.Name+".csv"), kline)
						return nil
					}, retry)
					logs.PrintErr(err)
				}
				return nil
			})
		}(codes[i])

	}

	ch.Wait()

	return nil
}

func toCsv(c *tdx.Client, filename string, kline v1.Klines) error {

	data := [][]any{
		{"日期", "代码", "名称", "昨收", "今开", "最高", "最低", "现收", "总手", "金额", "涨幅", "涨幅比"},
	}
	for _, k := range kline {
		data = append(data, []any{
			time.Unix(k.Unix, 0).Format(time.DateTime), k.Exchange + k.Code, c.Code.GetName(k.Exchange + k.Code),
			0, k.Open, k.High, k.Low, k.Close, k.Volume, k.Amount, k.RisePrice, k.RiseRate,
		})
	}

	buf, err := csv.Export(data)
	if err != nil {
		return err
	}

	return oss.New(filename, buf)

}
