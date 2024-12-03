package tdx

import (
	"github.com/injoyai/goutil/times"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/stock/data/tdx/model"
	"github.com/injoyai/tdx"
	"time"
)

func NewReal(hosts []string, op ...client.Option) (*Real, error) {
	c, err := tdx.DialWith(tdx.NewHostDial(hosts), op...)
	if err != nil {
		return nil, err
	}
	return &Real{
		Client: c,
	}, nil
}

type Real struct {
	*tdx.Client
}

func (this *Real) Get(code string, cache model.Klines) (model.Klines, error) {

	last := &model.Kline{Unix: times.IntegerDay(time.Now()).Unix()}
	if len(cache) > 0 {
		last = cache[len(cache)-1]   //获取最后的数据,用于截止获取数据
		cache = cache[len(cache)-1:] //删除最后一分钟的数据,最后一分钟实时统计的,用新数据更新
	}

	size := uint16(800)
	list := model.Klines(nil)
	for {
		resp, err := this.Client.GetKlineMinute(code, 0, size)
		if err != nil {
			return cache, err
		}

		done := false
		for _, v := range resp.List {
			//获取今天有效的分时图
			if last.Unix <= v.Time.Unix() {
				list = append(list, model.NewKline(code, v, v.Time))
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
