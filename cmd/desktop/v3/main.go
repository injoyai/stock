package main

import (
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/tray"
	"github.com/injoyai/goutil/oss/win"
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/cmd/internal/chart"
	"github.com/injoyai/stock/data/tdx/model"
	"github.com/injoyai/stock/data/tdx/v2"
	"github.com/injoyai/stock/util/csv"
	"github.com/injoyai/stock/util/zip"
	"github.com/robfig/cron/v3"
	"path/filepath"
	"strings"
	"time"
)

func init() {
	logs.SetShowColor(false)
	cfg.Init(
		func() conv.IGetVar {
			execDir := oss.ExecDir()
			switch {
			case strings.HasPrefix(execDir, "C:\\Users") && !strings.HasSuffix(execDir, "\\Start Menu\\Programs\\Startup"):
				return cfg.WithFile("./config/config.yaml")
			}
			return cfg.WithFile(filepath.Join(execDir, "/config/config.yaml"))
		}(),
		cfg.WithFlag(
			&cfg.Flag{Name: "hosts", Usage: "服务器地址"},
			&cfg.Flag{Name: "number", Usage: "客户端数量"},
			&cfg.Flag{Name: "limit", Usage: "协程数量"},
			&cfg.Flag{Name: "database", Usage: "数据存储位置"},
			&cfg.Flag{Name: "codes", Usage: "爬取的股票代码(sz000001)"},
			&cfg.Flag{Name: "runFirst", Usage: "启动立马运行"},
		),
	)
}

func main() {

	conf := &tdx.Config{
		Hosts:    cfg.GetStrings("hosts"),
		Number:   cfg.GetInt("number", 10),
		Limit:    cfg.GetInt("limit", 20),
		Database: cfg.GetString("database"),
		Queue: tdx.QueueConfig{
			OpenCap:   cfg.GetInt("queue.openCap", 100),
			OpenLimit: cfg.GetInt("queue.openLimit", 100),
			PullCap:   cfg.GetInt("queue.pullCap", 100),
			PullLimit: cfg.GetInt("queue.pullLimit", 100),
			SaveCap:   cfg.GetInt("queue.saveCap", 100),
			SaveLimit: cfg.GetInt("queue.saveLimit", 100),
			Retry:     cfg.GetInt("queue.retry", 3)},
	}

	tray.Run(
		func(s *tray.Stray) {
			s.SetIco(IcoStock)
			s.AddMenu().SetName("版本: v0.3.1").Disable()
			last := s.AddMenu().SetName("上次:").Disable()
			next := s.AddMenu().SetName("下次:").Disable()
			start := s.AddMenu().SetName("执行")
			go func() {
				task := cron.New(cron.WithSeconds())
				task.Start()
				taskid := cron.EntryID(0)

				//连接客户端
				c, err := tdx.Dial(conf, tdx.WithDebug(false))
				logs.PanicErr(err)

				f := func(up bool) {
					defer func() {
						last.SetName(time.Now().Format("上次: 01-02 15:04"))
						next.SetName(task.Entry(taskid).Next.Format("下次: 01-02 15:04"))
					}()
					if up {
						codes := cfg.GetStrings("codes", c.Code.GetStocks())
						start.Disable().SetName("执行中...")
						msg := fmt.Sprintf("开始更新数据,数量[%d]", len(codes))
						logs.Info(msg)
						notice.DefaultWindows.Publish(&notice.Message{Title: "Stock Desktop", Content: msg})
						defer func() {
							start.SetName("执行").Enable()
							logs.Info("数据更新完成...")
							notice.DefaultWindows.Publish(&notice.Message{Title: "Stock Desktop", Content: "数据更新完成..."})
						}()
						logs.PrintErr(update(s, c, codes))
					}

				}
				start.OnClick(func(m *tray.Menu) { f(true) })

				//每天下午16点进行数据更新
				taskid, _ = task.AddFunc("0 0 16 * * *", func() { f(c.Workday.TodayIs()) })

				//更新数据
				f(cfg.GetBool("runFirst", true))

				<-s.Done()
			}()
		},
		WithChart(),
		WithStartup(),
		tray.WithSeparator(),
		tray.WithExit(),
		tray.WithHint("定时拉取股票信息"),
	)

}

func update(s *tray.Stray, c *tdx.Client, codes []string) error {

	plan := NewPlan(len(codes))
	s.SetHint(plan.String())

	c.UpdateWait(codes, func(data *tdx.PullDataAll) {
		for _, v := range data.Data {
			toCsv(c.Code, filepath.Join(c.Cfg.Database, "csv", data.DB.Code, tdx.HandlerMap[v.Table].Name+".csv"), v.Klines())
		}
		s.SetHint(plan.Add().String())
	})

	logs.Info("开始压缩数据...")

	//进行压缩操作,250ms
	s.SetHint(plan.CompressStart().String())
	err := zip.Encode(filepath.Join(c.Cfg.Database, "csv")+"/", filepath.Join(c.Cfg.Database, "csv.zip"))
	logs.PrintErr(err)
	s.SetHint(plan.CompressEnd().String())

	return nil
}

func toCsv(c *tdx.Code, filename string, kline model.Klines) error {

	data := [][]any{
		{"日期", "代码", "名称", "昨收", "今开", "最高", "最低", "现收", "总手", "金额", "涨幅", "涨幅比"},
	}
	for _, k := range kline {
		data = append(data, []any{
			time.Unix(k.Unix, 0).Format(time.DateTime), k.Exchange + k.Code, c.GetName(k.Exchange + k.Code),
			k.Last, k.Open, k.High, k.Low, k.Close, k.Volume, k.Amount, k.RisePrice, k.RiseRate,
		})
	}

	buf, err := csv.Export(data)
	if err != nil {
		return err
	}

	return oss.New(filename, buf)

}

func WithChart() tray.Option {
	return func(s *tray.Stray) {
		s.AddMenu().SetName("实时").OnClick(func(m *tray.Menu) {
			chart.Show()
		})
	}
}

func WithStartup() tray.Option {
	return func(s *tray.Stray) {
		filename := oss.ExecName()
		_, name := filepath.Split(filename)
		name = strings.Split(name, ".")[0]
		startupFilename := oss.UserStartupDir(name + ".lnk")
		s.AddMenuCheck().SetChecked(oss.Exists(startupFilename)).
			SetName("自启").OnClick(func(m *tray.Menu) {
			if !m.Checked() {
				logs.PrintErr(win.CreateStartupShortcut(filename))
				m.Check()
			} else {
				logs.PrintErr(oss.Remove(startupFilename))
				m.Uncheck()
			}
		})
	}
}
