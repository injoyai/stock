package tdx

import (
	"errors"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/tdx"
)

func NewPool(hosts []string, cap int, op ...client.Option) (*Pool, error) {
	if cap <= 0 {
		cap = 1
	}
	p := &Pool{
		ch:     make(chan *tdx.Client, cap),
		all:    make([]*tdx.Client, cap),
		Closer: safe.NewCloser(),
	}
	p.Closer.SetCloseFunc(func(err error) error {
		for _, v := range p.all {
			v.CloseAll()
		}
		return nil
	})
	for i := 0; i < cap; i++ {
		c, err := tdx.DialWith(tdx.NewHostDial(hosts, 0), op...)
		if err != nil {
			p.Close()
			return nil, err
		}
		c.SetRedial()
		p.all[i] = c
		p.ch <- c
	}
	return p, nil
}

type Pool struct {
	ch  chan *tdx.Client
	all []*tdx.Client
	*safe.Closer
}

func (this *Pool) Get() *tdx.Client {
	c := <-this.ch
	return c
}

func (this *Pool) Get2() (*tdx.Client, error) {
	select {
	case <-this.Done():
		return nil, this.Err()
	case c, ok := <-this.ch:
		if !ok {
			return nil, errors.New("已关闭")
		}
		return c, nil
	}
}

func (this *Pool) Put(c *tdx.Client) {
	select {
	case <-this.Done():
		return
	case this.ch <- c:
	}
}

func (this *Pool) Retry(f func(c *tdx.Client) error, retry int) error {
	c, err := this.Get2()
	if err != nil {
		return err
	}
	defer this.Put(c)
	return g.Retry(func() error { return f(c) }, retry)
}
