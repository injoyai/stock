package main

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/common"
	"github.com/injoyai/stock/data/tdx"
)

func main() {

	//连接客户端
	c, err := tdx.Dial(tdx.Hosts)
	logs.PanicErr(err)

	//更新数据
	logs.PrintErr(c.UpdateCodes(c.GetCodes()))

	//每天下午16点进行数据更新
	common.Corn.SetTask("update", "0 0 16 * * *", func() {
		logs.PrintErr(c.UpdateCodes(c.GetCodes()))
	})

	//等待客户端退出
	<-c.Client.Done()
}
