package main

import (
	_ "embed"
	"github.com/injoyai/stock/cmd/internal/chart"
)

func main() {
	chart.Show()
}
