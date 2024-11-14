package main

import (
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/logs"
	"github.com/injoyai/lorca"
	"github.com/injoyai/stock/data/tdx"
	"github.com/injoyai/stock/gui"
	"github.com/injoyai/tdx/protocol"
	"time"
)

func main() {

	code := ""

	for {
		code = g.Input("请输入代码(例sz000001):")
		_, _, err := protocol.DecodeCode(code)
		if err != nil {
			logs.Err(err)
		} else {
			break
		}
	}

	//连接客户端
	c, err := tdx.Dial("124.71.187.122", func(c *client.Client) {
		c.Logger.Debug(false)
	})
	logs.PanicErr(err)

	lorca.Run(&lorca.Config{
		Width:  700,
		Height: 400,
		Html:   gui.ChartHtml,
	}, func(app lorca.APP) error {

		quote, err := c.Quote(code)
		if err != nil {
			return err
		}

		for range g.Interval(time.Second * 2) {
			select {
			case <-app.Done():
				return nil

			default:
				ls, err := c.KlineReal(code, nil)
				if err != nil {
					logs.Err(err)
					continue
				}
				data := ls.ChartDay(quote.K.Last.Float64(), c.GetCodeName(code))
				data.Init()
				app.Eval(fmt.Sprintf("loading(%s,%f,%f)", conv.String(data), data.Min, data.Max))

			}
		}

		return nil
	})
}
