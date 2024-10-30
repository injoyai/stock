package main

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/common"
	"github.com/injoyai/stock/data/tdx"
)

func main() {
	logs.PanicErr(common.Init())

	logs.PanicErr(tdx.Init())
}
