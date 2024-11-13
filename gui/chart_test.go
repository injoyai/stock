package gui

import "testing"

func TestShowChart(t *testing.T) {
	ShowChart(&Chart{
		Labels: []string{"1", "2", "3", "4", "5"},
		Datasets: []*ChartItem{{
			Label: "曲线1",
			Data:  []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			Color: "rgba(75, 192, 192, 1)",
		}},
	})
}
