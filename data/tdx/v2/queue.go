package tdx

import (
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/data/tdx/model"
	"github.com/injoyai/tdx/protocol"
	"time"
	"xorm.io/xorm"
)

func NewQueue(cap, limit int) *Queue {
	return &Queue{
		ch:    make(chan QueueFunc, cap),
		limit: make(chan struct{}, limit),
		retry: 3,
	}
}

type Queue struct {
	ch    chan QueueFunc
	limit chan struct{}
	retry int
	next  chan func()
}

func (this *Queue) Run() {
	for {
		f := <-this.ch
		this.limit <- struct{}{}
		go func() {
			defer func() { <-this.limit }()
			g.Retry(func() error {
				cache, err := f.GetCache()
				if err != nil {
					return err
				}
				this.next <- func() {
					xx := &Klines{
						DB: f.GetDB(),
					}
					for _, ls := range cache {
						k, err := f.CacheUpdate(ls, nil, nil)
						logs.PrintErr(err)
						xx.Data = append(xx.Data, k)
					}

				}
				return nil
			}, this.retry)
		}()

	}
}

func (this *Queue) Add(f QueueFunc) {
	this.ch <- f
}

func (this *Queue) RunNext() {
	for {
		f := <-this.next
		f()
	}
}

type QueueFunc interface {
	GetDB() *xorms.Engine
	GetCache() (Cache, error)
	CacheUpdate(cache []*model.Kline,
		get func(code string, start, count uint16) (*protocol.KlineResp, error),
		dealTime func(t time.Time) time.Time) (*Kline, error)
}

type Cache map[string][]*model.Kline

type Klines struct {
	DB   *xorms.Engine
	Data []*Kline
}

func (this *Klines) Update() {
	for _, v := range this.Data {
		if err := v.update(this.DB); err != nil {
			logs.Error(err)
		}
	}
}

type Kline struct {
	Table  string         //表名
	Cache  []*model.Kline //数据库中的数据
	Update *model.Kline   //需要更新的数据
	Insert []*model.Kline //拉取过来的数据
}

func (this *Kline) update(db *xorms.Engine) error {
	//4. 将缺的数据入库
	return db.SessionFunc(func(session *xorm.Session) error {
		if _, err := session.Table(this.Table).Insert(this.Update); err != nil {
			return err
		}
		for _, v := range this.Insert {
			if _, err := session.Table(this.Table).Where("Node=?", v.Node).Update(v); err != nil {
				return err
			}
		}
		return nil
	})
}
