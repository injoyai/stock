package main

import (
	_ "embed"
	"errors"
	"time"

	"github.com/injoyai/logs"
	"github.com/injoyai/lorca"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
)

//go:embed kline.html
var klineHtml string

func main() {

	c := NewClient()

	lorca.Run(&lorca.Config{
		Html:   klineHtml,
		Width:  640, //  1024,
		Height: 480, // 768,
	}, func(app lorca.APP) error {

		app.Bind("getKlineDay", func(code string, start, count uint16) any {
			logs.Debug("请求K线数据:", code, start, count)

			ls, err := c.KlineDay(code, start, count)
			if err != nil {
				logs.Err("获取K线数据错误:", err)
				return []interface{}{}
			}

			if ls == nil || len(ls) == 0 {
				logs.Warn("未获取到K线数据")
				return []interface{}{}
			}

			// 转换数据格式
			result := make([][]float64, 0, len(ls))
			for _, item := range ls {
				if item == nil {
					continue
				}
				// 直接构造数组格式，符合前端预期
				result = append(result, []float64{
					float64(item.Time.Unix()), // 时间戳
					item.Open.Float64(),       // 开盘价
					item.Close.Float64(),      // 收盘价
					item.Low.Float64(),        // 最低价
					item.High.Float64(),       // 最高价
					float64(item.Volume),      // 成交量
				})
			}

			if len(result) == 0 {
				return []interface{}{}
			}

			return result
		})

		app.Bind("getQuote", func(code string) any {

			ls, err := c.Quote(code)
			if err != nil {
				logs.Err(err)
				return map[string]string{"error": err.Error()}
			}
			return ls

		})

		return nil
	})

}

func NewClient() *Client {
	cli := &Client{}

	c, err := tdx.DialHosts(tdx.SHHosts, tdx.WithRedial())
	if err != nil {
		logs.Err(err)
		go func() {
			for ; err != nil; <-time.After(time.Second * 5) {
				c, err = tdx.DialHosts(tdx.SHHosts, tdx.WithRedial())
			}
			cli.Client = c
		}()
	}
	cli.Client = c
	return cli
}

type Client struct {
	*tdx.Client
}

func (this *Client) Quote(code string) (*protocol.Quote, error) {
	if this.Client == nil || this.Client.Closed() {
		return nil, errors.New("客户端连接失败")
	}
	resp, err := this.Client.GetQuote(code)
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 {
		return nil, errors.New("not found")
	}
	return resp[0], nil
}

func (this *Client) KlineDay(code string, start, count uint16) ([]*protocol.Kline, error) {
	if this.Client == nil || this.Client.Closed() {
		return nil, errors.New("客户端连接失败")
	}
	resp, err := this.Client.GetKlineDay(code, start, count)
	if err != nil {
		return nil, err
	}
	return resp.List, nil
}
