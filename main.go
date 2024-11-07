package main

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/api"
	"github.com/injoyai/stock/data/tdx"
)

func main() {
	//加载tdx数据爬取服务
	//logs.PanicErr(tdx.Init())

	////更新代码
	//logs.PanicErr(tdx.UpdateCode(false))

	//更新分时图
	//logs.PanicErr(tdx.GetStockHistoryKline())

	//每天凌晨进行数据更新
	//common.Corn.SetTask("updateCode", "0 0 0 * * *", func() {
	//
	//})

	//每秒更新实时数据
	//common.Corn.SetTask("updateReal", "* * * * * *", func() {
	//
	//})

	c, err := tdx.Dial("124.71.187.122")
	logs.PanicErr(err)

	_, err = c.KlineDay("sz000001")
	logs.PanicErr(err)

	//加载http服务
	logs.Err(api.Run())
}
