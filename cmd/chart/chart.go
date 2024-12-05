package main

import (
	_ "embed"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/logs"
	"github.com/injoyai/lorca"
	"github.com/injoyai/stock/data/tdx"
	"github.com/injoyai/stock/data/tdx/model"
	tdx2 "github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
	"math"
	"time"
)

//go:embed chart.html
var ChartHtml string

var Colors = []string{"rgba(75, 192, 192)", "rgba(192, 75, 75)", "rgb(255, 99, 132)", "rgb(54, 162, 235)", "rgb(255, 206, 86)", "rgb(75, 192, 192)", "rgb(153, 102, 255)", "rgb(255, 159, 64)"}

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
	c, err := tdx.Dial(&tdx.Config{}, func(c *client.Client) {
		c.Logger.Debug(false)
	})
	logs.PanicErr(err)

	lorca.Run(&lorca.Config{
		Width:  700,
		Height: 400,
		Html:   ChartHtml,
	}, func(app lorca.APP) error {

		return c.WithOpenDB(code, func(db *tdx.DB) error {

			return c.Pool.Retry(func(cli *tdx2.Client) error {
				quote, err := db.Quote(cli)
				if err != nil {
					return err
				}

				for ; ; <-time.After(time.Second * 2) {
					select {
					case <-app.Done():
						return nil

					default:
						ls, err := c.Real.Get(code, nil)
						if err != nil {
							logs.Err(err)
							continue
						}

						data := ChartDay(ls, quote.K.Last.Float64(), c.Code.GetName(code))
						data.Init()
						app.Eval(fmt.Sprintf("loading(%s,%f,%f)", conv.String(data), data.Min, data.Max))

					}
				}

			}, 1)

			return nil
		})

	})
}

func ChartDay(ls []*model.Kline, last float64, name string) *Chart {
	dayMinute := 60 * 4
	c := &Chart{
		Labels: make([]string, dayMinute),
		Datasets: []*ChartItem{{
			Label: name,
			Data:  make([]float64, len(ls)),
		}},
	}

	now := time.Date(2024, 1, 1, 9, 31, 0, 0, time.Local)
	for i := 0; i < dayMinute/2; i++ {
		c.Labels[i] = now.Add(time.Minute * time.Duration(i)).Format("15:04")
	}

	now = time.Date(2024, 1, 1, 13, 0, 0, 0, time.Local)
	for i := 0; i < dayMinute/2; i++ {
		c.Labels[i+dayMinute/2] = now.Add(time.Minute * time.Duration(i)).Format("15:04")
	}

	var sub float64
	for i, v := range ls {
		c.Datasets[0].Data[i] = v.Close
		val := math.Abs(v.Close - last)
		if val > sub {
			sub = val
		}
	}

	c.Max = (last + sub) * 1.02
	c.Min = (last - sub) * 0.98

	return c
}
