package tdx

import (
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/tdx"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type QueueConfig struct {
	OpenCap   int
	OpenLimit int
	PullCap   int
	PullLimit int
	SaveCap   int
	SaveLimit int
	Retry     int
}

type Config struct {
	Hosts    []string //服务端IP
	Number   int      //客户端数量
	Limit    int      //协程数量
	Database string   //数据位置
	Queue    QueueConfig
}

func (this *Config) init() *Config {
	if len(this.Hosts) == 0 {
		this.Hosts = tdx.Hosts
	}
	if len(this.Database) == 0 {
		execDir := oss.ExecDir()
		switch {
		case strings.HasPrefix(execDir, "C:\\Users") && !strings.HasSuffix(execDir, "\\Start Menu\\Programs\\Startup"):
			//默认IDE缓存的地方,则读取代码位置的配置
			//C:\Users\Admin\AppData\Local\JetBrains\GoLand2024.1\tmp\GoLand\___1go_build_github_com_injoyai_stock_cmd_desktop.exe
			this.Database = "./database/"
		default:
			this.Database = filepath.Join(oss.ExecDir(), "/database")
		}
	}
	if this.Number <= 0 {
		this.Number = 1
	}
	if this.Queue.OpenLimit <= 0 {
		this.Queue.OpenLimit = 100
	}
	if this.Queue.OpenCap <= 0 {
		this.Queue.OpenCap = 100
	}
	if this.Queue.PullCap <= 0 {
		this.Queue.PullCap = 100
	}
	if this.Queue.PullLimit <= 0 {
		this.Queue.PullLimit = 10
	}
	if this.Queue.SaveLimit <= 0 {
		this.Queue.SaveLimit = 100
	}
	if this.Queue.SaveCap <= 0 {
		this.Queue.SaveCap = 100
	}
	if this.Queue.Retry <= 0 {
		this.Queue.Retry = 3
	}

	os.Mkdir(this.Database, os.ModePerm)
	return this
}

func WithDebug(b ...bool) client.Option {
	return func(c *client.Client) {
		c.Logger.Debug(b...)
	}
}

func Dial(cfg *Config, op ...client.Option) (cli *Client, err error) {

	cli = &Client{Cfg: cfg.init()}

	option := func(c *client.Client) {
		c.Logger.Debug()
		c.SetRedial(true)
		c.SetOption(op...)
	}

	{ //工作日信息
		cli.Workday, err = NewWorkday(cfg.Hosts, filepath.Join(cfg.Database, "workday.db"), option)
		if err != nil {
			return nil, err
		}
	}

	{ //股票代码
		cli.Code, err = NewCode(cfg.Hosts, filepath.Join(cfg.Database, "codes.db"), option)
		if err != nil {
			return nil, err
		}
	}

	{ //实时价格
		cli.Real, err = NewReal(cfg.Hosts, option)
		if err != nil {
			return nil, err
		}
	}

	{ //更新队列
		pool, err := NewPool(cfg.Hosts, cfg.Number, option)
		if err != nil {
			return nil, err
		}
		cli.update = NewUpdate(pool, cfg.Queue.OpenCap, cfg.Queue.OpenLimit, cfg.Queue.PullCap, cfg.Queue.PullLimit, cfg.Queue.SaveCap, cfg.Queue.SaveLimit, cfg.Queue.Retry)
	}

	return cli, nil
}

type Client struct {
	Cfg       *Config  //配置信息
	Code      *Code    //股票代码
	Workday   *workday //工作日
	Real      *Real    //实时价格
	update    *Update  //更新实例
	OnUpdated func()   //完成更新事件
}

func (this *Client) Update(codes []string, onSaved func(code *PullDataAll)) {
	this.update.OnSaved = onSaved
	for _, code := range codes {
		this.update.Add(func() (*DBData, error) {
			db, err := NewDB(this.Cfg.Database, code)
			if err != nil {
				return nil, err
			}
			return db.GetCache()
		})
	}
}

func (this *Client) UpdateWait(codes []string, onSaved func(code *PullDataAll)) {
	wg := &sync.WaitGroup{}
	wg.Add(len(codes))
	this.Update(codes, func(code *PullDataAll) {
		onSaved(code)
		wg.Done()
	})
	wg.Wait()
}
