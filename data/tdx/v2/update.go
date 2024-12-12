package tdx

import (
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/data/tdx/model"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
	"time"
	"xorm.io/xorm"
)

func NewUpdate(pool *Pool, openCap, openLimit, pullCap, pullLimit, saveCap, saveLimit, retry int) *Update {
	up := &Update{
		openChan:  make(chan func() (*DBData, error), openCap),
		openLimit: make(chan struct{}, openLimit),
		pullChan:  make(chan *DBData, pullCap),
		pullLimit: make(chan struct{}, pullLimit),
		saveChan:  make(chan *PullDataAll, saveCap),
		saveLimit: make(chan struct{}, saveLimit),
	}
	go up.runOpen(retry)
	go up.runPull(pool, retry)
	go up.runSave(retry)
	return up
}

type Update struct {

	//1. 打开数据库,读取数据,加入队列
	//2. 拉取数据,加入队列
	//3. 写入数据库,并关闭

	openChan  chan func() (*DBData, error)
	openLimit chan struct{}

	pullChan  chan *DBData
	pullLimit chan struct{}

	saveChan  chan *PullDataAll
	saveLimit chan struct{}

	OnSaved func(data *PullDataAll)
}

func (this *Update) Add(f func() (*DBData, error)) {
	this.openChan <- f
}

func (this *Update) runOpen(retry int) {
	for {
		f := <-this.openChan
		this.openLimit <- struct{}{}

		go func(f func() (*DBData, error)) {
			defer func() { <-this.openLimit }()
			g.Retry(func() error {
				data, err := f()
				if err != nil {
					return err
				}
				this.pullChan <- data
				return nil
			}, retry)
		}(f)
	}
}

func (this *Update) runPull(pool *Pool, retry int) {
	for {
		data := <-this.pullChan
		this.pullLimit <- struct{}{}
		go func(data *DBData) {
			defer func() { <-this.pullLimit }()
			c, err := pool.Get()
			if err != nil {
				logs.Err(err)
				return
			}
			pullData := data.PullAll(c, retry)
			pool.Put(c)
			this.saveChan <- pullData
		}(data)
	}
}

func (this *Update) runSave(retry int) {
	for {
		data := <-this.saveChan
		this.saveLimit <- struct{}{}
		go func(data *PullDataAll) {
			defer func() {
				<-this.saveLimit
				if this.OnSaved != nil {
					this.OnSaved(data)
				}
			}()
			data.SaveAndClose(retry)
		}(data)
	}
}

type DBData struct {
	DB   *DB                       //实例
	Data map[string][]*model.Kline //缓存的数据
}

func (this *DBData) Code() string {
	return this.DB.Code
}

func (this *DBData) PullAll(c *tdx.Client, retry int) *PullDataAll {
	result := &PullDataAll{
		DB: this.DB,
	}
	for table, v := range this.Data {
		g.Retry(func() error {
			data, err := this.Pull(c, table, v, HandlerMap[table].Handler, HandlerMap[table].Node)
			if err != nil {
				return err
			}
			result.Data = append(result.Data, data)
			return nil
		}, retry)
	}
	return result
}

func (this *DBData) Pull(c *tdx.Client, table string, cache []*model.Kline,
	pull func(c *tdx.Client, code string, start, count uint16) (*protocol.KlineResp, error),
	dealTime func(t time.Time) time.Time) (*PullData, error) {
	last := new(model.Kline)
	if len(cache) > 0 {
		last = cache[len(cache)-1]   //获取最后一条数据,用于截止从服务器拉的数据
		cache = cache[:len(cache)-1] //去除最后一条数据,用拉取过来的数据更新掉
	}

	//3. 从服务器拉取数据
	list := []*model.Kline(nil)
	size := uint16(800)
	size = 8
	for start := uint16(0); ; start += size {
		resp, err := pull(c, this.Code(), start, size)
		if err != nil {
			logs.Err(err)
			return nil, err
		}

		done := false
		ls := []*model.Kline(nil)
		for _, v := range resp.List {
			node := dealTime(v.Time)
			if last.Node <= node.Unix() {
				ls = append(ls, model.NewKline(this.Code(), v, node))
			} else {
				done = true
			}
		}
		list = append(ls, list...)
		if resp.Count < size || done {
			break
		}

		size *= 10
		if size > 800 {
			size = 800
		}
	}
	return &PullData{
		Table:  table,
		Cache:  cache,
		Update: list[0],
		Insert: list[1:],
	}, nil
}

type PullDataAll struct {
	DB   *DB
	Data []*PullData
}

func (this *PullDataAll) SaveAndClose(retry int) {
	for _, v := range this.Data {
		g.Retry(func() error { return v.save(this.DB.db) }, retry)
	}
	this.DB.Close()
}

type PullData struct {
	Table  string         //表名
	Cache  []*model.Kline //数据库中的数据
	Update *model.Kline   //需要更新的数据
	Insert []*model.Kline //拉取过来的数据
}

func (this *PullData) Klines() []*model.Kline {
	cache := append(this.Cache, this.Update)
	cache = append(cache, this.Insert...)
	return cache
}

func (this *PullData) save(db *xorms.Engine) error {
	//4. 将缺的数据入库
	return db.SessionFunc(func(session *xorm.Session) error {
		if _, err := session.Table(this.Table).Where("Node=?", this.Update.Node).Update(this.Update); err != nil {
			return err
		}
		for _, v := range this.Insert {
			if _, err := session.Table(this.Table).Insert(v); err != nil {
				return err
			}
		}
		return nil
	})
}
