package main

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/api"
	"github.com/injoyai/stock/data/tdx"
)

func main() {
	//加载tdx数据爬取服务
	logs.PanicErr(tdx.Init())

	//加载http服务
	logs.Err(api.Run())
}
