package tdx

import (
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/goutil/times"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
	"github.com/robfig/cron/v3"
	"os"
	"time"
	"xorm.io/xorm"
)

var (
	Hosts = tdx.Hosts
)

type Config struct {
	Hosts    []string
	Cap      int
	Database string
	Workday  string
}

type Cli = tdx.Client

func Dial(cfg *Config, op ...client.Option) (*Client, error) {

	if len(cfg.Hosts) == 0 {
		cfg.Hosts = Hosts
	}
	if len(cfg.Database) == 0 {
		cfg.Database = "./database/"
	}
	if cfg.Cap <= 0 {
		cfg.Cap = 1
	}
	if len(cfg.Workday) == 0 {
		cfg.Workday = "workday"
	}
	os.Mkdir(cfg.Database, os.ModePerm)

	cli := &Client{Cfg: cfg}

	//增加一个自用客户端,用于获取工作日信息
	c, err := tdx.DialWith(tdx.NewHostDial(cfg.Hosts), op...)
	if err != nil {
		return nil, err
	}
	db, err := cli.OpenDB(cfg.Workday)
	if err != nil {
		return nil, err
	}
	cli.Workday, err = newWorkday(c, db)
	if err != nil {
		logs.Err(err)
		return nil, err
	}

	//新建连接池,
	cli.Pool, err = NewPool(cfg.Hosts, cfg.Cap, func(c *client.Client) {
		c.Logger.Debug()
		c.SetRedial(true)
		c.SetOption(op...)
	})
	if err != nil {
		return nil, err
	}

	update := func(must bool) error {
		//1. 更新工作日数据
		logs.Debug("更新: 工作日数据")
		err = cli.Workday.Update()
		logs.PrintErr(err)
		//2. 更新代码信息
		logs.Debug("更新: 代码信息")
		return cli.UpdateCode(!must && !cli.Workday.TodayIs())
	}

	//启动更新一次
	if err := update(true); err != nil {
		cli.Pool.Close()
		return nil, err
	}

	//每天8点更新代码信息,比如新增了股票,或者股票改了名字
	cron.New(cron.WithSeconds()).AddFunc("0 0 8 * * *", func() {
		logs.PrintErr(update(false))
	})

	return cli, nil
}

/*
Client 客户端
*/
type Client struct {
	Cfg     *Config
	Pool    *Pool
	Codes   map[string]*Code
	codeDB  *xorms.Engine //代码数据库实例
	Workday *workday      //工作日
}

func (this *Client) WithOpenDB(code string, f func(db *DB) error) error {
	db, err := NewDB(this.Cfg.Database, code)
	if err != nil {
		return err
	}
	defer db.Close()
	return f(db)
}

// UpdateCodes 更新股票
func (this *Client) UpdateCodes(codes []string, retrys ...int) error {
	retry := conv.DefaultInt(3, retrys...)

	//2. 遍历全部股票
	for i := 0; i < len(codes); i++ {
		err := this.WithOpenDB(codes[i], func(db *DB) error {
			return this.Pool.Retry(func(c *tdx.Client) error { return db.Update(c) }, retry)
		})
		logs.PrintErr(err)

	}

	return nil
}

// GetStockCodes 获取股票代码,不一定全
func (this *Client) GetStockCodes() []string {
	ls := []string(nil)
	for k, _ := range this.Codes {
		if len(k) == 8 {
			switch k[:2] {
			case "sz":
				if IsStock(protocol.ExchangeSZ, k[2:]) {
					ls = append(ls, k)
				}
			case "sh":
				if IsStock(protocol.ExchangeSH, k[2:]) {
					ls = append(ls, k)
				}
			}
		}
	}
	return ls
}

// GetCodeName 获取股票中文名称
func (this *Client) GetCodeName(code string) string {
	if v, ok := this.Codes[code]; ok {
		return v.Name
	}
	return "未知"
}

// OpenDB 打开数据库,内部使用
func (this *Client) OpenDB(code string, entity ...any) (*xorms.Engine, error) {
	filename := this.Cfg.Database + code + ".db"
	db, err := sqlite.NewXorm(filename)
	if err != nil {
		return nil, err
	}
	if err := db.Sync(entity...); err != nil {
		return nil, err
	}
	return db, nil
}

func (this *Client) UpdateCode(byDatabase bool) error {
	codes, err := this.Code(byDatabase)
	if err != nil {
		return err
	}
	codeMap := make(map[string]*Code)
	for _, code := range codes {
		codeMap[code.Exchange+code.Code] = code
	}
	this.Codes = codeMap
	return nil
}

// Code 更新股票并返回结果
func (this *Client) Code(byDatabase bool) ([]*Code, error) {
	logs.Debug("更新代码信息")

	c, err := this.Pool.Get()
	if err != nil {
		return nil, err
	}
	defer this.Pool.Put(c)

	//1. 打开数据库
	db, err := this.OpenDB("code", new(Code))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	//2. 查询数据库所有股票
	list := []*Code(nil)
	if err := db.Find(&list); err != nil {
		return nil, err
	}

	//如果是从缓存读取,则返回结果
	if byDatabase {
		return list, nil
	}

	mCode := make(map[string]*Code, len(list))
	for _, v := range list {
		mCode[v.Code] = v
	}

	//3. 从服务器获取所有股票代码
	insert := []*Code(nil)
	update := []*Code(nil)
	for _, exchange := range []protocol.Exchange{protocol.ExchangeSH, protocol.ExchangeSZ} {
		resp, err := c.GetCodeAll(exchange)
		if err != nil {
			return nil, err
		}
		for _, v := range resp.List {
			if _, ok := mCode[v.Code]; ok {
				if mCode[v.Code].Name != v.Name {
					mCode[v.Code].Name = v.Name
					update = append(update, NewCode(exchange, v))
				}
			} else {
				code := NewCode(exchange, v)
				insert = append(insert, code)
				list = append(list, code)
			}
		}
	}

	//4. 插入或者更新数据库
	err = db.SessionFunc(func(session *xorm.Session) error {
		for _, v := range insert {
			if _, err := session.Insert(v); err != nil {
				return err
			}
		}
		for _, v := range update {
			if _, err := session.Where("Code=?", v.Code).Cols("Name").Update(v); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return list, nil

}

// KlineReal 分钟实时获取
func (this *Client) KlineReal(code string, cache Klines) (Klines, error) {

	c, err := this.Pool.Get()
	if err != nil {
		return cache, err
	}
	defer this.Pool.Put(c)

	last := &Kline{Unix: times.IntegerDay(time.Now()).Unix()}
	if len(cache) > 0 {
		last = cache[len(cache)-1]   //获取最后的数据,用于截止获取数据
		cache = cache[len(cache)-1:] //删除最后一分钟的数据,最后一分钟实时统计的,用新数据更新
	}

	size := uint16(800)
	list := Klines(nil)
	for {
		resp, err := c.GetKlineMinute(code, 0, size)
		if err != nil {
			return cache, err
		}

		done := false
		for _, v := range resp.List {
			//获取今天有效的分时图
			if last.Unix <= v.Time.Unix() {
				list = append(list, NewKline(code, v))
			} else {
				done = true
			}
		}

		if resp.Count < size || done {
			break
		}

	}

	cache = append(cache, list...)
	return cache, nil

}
