package main

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/logs"
	"github.com/injoyai/lorca"
	"github.com/injoyai/stock/data/tdx"
	"github.com/injoyai/stock/data/tdx/model"
	tdx2 "github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
	"math"
	"strings"
	"time"
)

//go:embed chart.html
var ChartHtml string

func main() {

	lorca.Run(&lorca.Config{
		Width:  300,
		Height: 200,
		Html:   ChartHtml,
	}, func(app lorca.APP) error {

		//连接客户端
		var c = &Client{}
		var err error
		for ; ; <-time.After(time.Second * 3) {
			c.Real, err = tdx.NewReal(tdx2.Hosts)
			if err == nil {
				break
			}
			app.Eval(fmt.Sprintf("notice('%s')", err.Error()))
		}

		return app.Bind("run", func() {

			err = func() error {

				code := app.GetValueByID("input")
				code = strings.ToLower(code)

				if len(code) != 8 || (code[:2] != "sh" && code[:2] != "sz") {
					return errors.New("股票代码不正确")
				}

				c.code = code
				quote, err := c.Quote()
				if err != nil {
					return err
				}

				b, _ := app.Bounds()
				b.Width = 800
				b.Height = 600
				app.SetBounds(b)

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

						data := ChartDay(ls, quote.K.Last.Float64(), code)
						data.Init()
						app.Eval(fmt.Sprintf("loading(%s,%f,%f)", conv.String(data), data.Min, data.Max))

					}
				}
			}()
			if err != nil {
				app.Eval(fmt.Sprintf("notice('%s')", err.Error()))
			}
		})

	})
}

type Client struct {
	code string
	*tdx.Real
}

func (this *Client) Quote() (*protocol.Quote, error) {
	resp, err := this.GetQuote(this.code)
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 {
		return nil, errors.New("not found")
	}
	return resp[0], nil
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
