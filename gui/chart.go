package gui

import (
	_ "embed"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/lorca"
)

//go:embed chart.html
var ChartHtml string

var (
	Colors = []string{"rgba(75, 192, 192)", "rgba(192, 75, 75)", "rgb(255, 99, 132)", "rgb(54, 162, 235)", "rgb(255, 206, 86)", "rgb(75, 192, 192)", "rgb(153, 102, 255)", "rgb(255, 159, 64)"}
)

func ShowChart(data *Chart) {

	lorca.Run(&lorca.Config{
		Width:  700,
		Height: 400,
		Html:   ChartHtml,
	}, func(app lorca.APP) error {
		data.Init()
		app.Eval(fmt.Sprintf("loading(%s,%f,%f)", conv.String(data), data.Min, data.Max))

		return nil
	})
}

type Chart struct {
	Max      float64      `json:"max"`
	Min      float64      `json:"min"`
	Labels   []string     `json:"labels"`
	Datasets []*ChartItem `json:"datasets"`
}

func (this *Chart) Init() {
	changeMax := this.Max == 0

	for i, v := range this.Datasets {
		v.init(i)
		if changeMax {
			for _, vv := range v.Data {
				if vv > this.Max {
					this.Max = vv
				}
			}
		}
	}

}

type ChartItem struct {
	Label       string    `json:"label"`
	Data        []float64 `json:"data"`
	Color       string    `json:"color"`
	BorderWidth int       `json:"borderWidth"`
	Tension     float64   `json:"tension"`
}

func (this *ChartItem) init(i int) {
	if len(this.Color) == 0 {
		this.Color = Colors[i%len(Colors)]
	}
	if this.BorderWidth == 0 {
		this.BorderWidth = 2
	}
	if this.Tension == 0 {
		this.Tension = 0.4
	}
}