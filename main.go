package main

import (
	"github.com/injoyai/goutil/g"
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

	//连接客户端
	c, err := tdx.Dial("124.71.187.122")
	logs.PanicErr(err)

	//每天下午16点进行数据更新
	//common.Corn.SetTask("updateCode", "0 0 16 * * *", func() {

	//1. 判断是否是节假日

	//2. 遍历全部股票
	for _, code := range []string{"sz000001"} {
		//3. 进行按股票进行每日更新,并尝试重试
		g.Retry(func() error {
			_, err = c.KlineDay(code)
			logs.PrintErr(err)
			return err
		}, 3)
	}

	//})

	//每秒更新实时数据
	//common.Corn.SetTask("updateReal", "* * * * * *", func() {
	//
	// 实时计算日策略
	//
	//})

	//加载http服务
	logs.Err(api.Run())
}
