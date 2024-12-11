package tdx

import (
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/tdx"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Hosts    []string //服务端IP
	Number   int      //客户端数量
	Limit    int      //协程数量
	Database string   //数据位置
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
	os.Mkdir(this.Database, os.ModePerm)
	return this
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

	{ //连接池
		cli.Pool, err = NewPool(cfg.Hosts, cfg.Number, option)
		if err != nil {
			return nil, err
		}
	}

	return cli, nil
}

type Client struct {
	Cfg       *Config  //配置信息
	Pool      *Pool    //连接池
	Code      *Code    //股票代码
	Workday   *workday //工作日
	Real      *Real    //实时价格
	Queue     *Queue   //
	OnUpdated func()   //完成更新时间
}

func (this *Client) WithOpenDB(code string, f func(db *DB) error) error {
	db, err := NewDB(this.Cfg.Database, code)
	if err != nil {
		return err
	}
	defer db.Close()
	return f(db)
}

func (this *Client) Update(codes []string) {

	for _, code := range codes {

		func() error {
			db, err := NewDB(this.Cfg.Database, code)
			if err != nil {
				return err
			}

			cache, err := db.GetCache()
			if err != nil {
				return err
			}

			this.Queue.Add(func() (*Cache, error) {
				cache, err := db.GetCache()
				if err != nil {
					return err
				}

			})

		}()

	}

}
