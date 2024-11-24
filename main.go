package main

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/common"
	"github.com/injoyai/stock/data/tdx"
)

func main() {

	//连接客户端
	c, err := tdx.Dial(&tdx.Config{Cap: 10, Database: "./database2/"})
	logs.PanicErr(err)

	codes := []string{"sz000001"}

	//更新数据
	logs.PrintErr(c.UpdateCodes(codes))

	//每天下午16点进行数据更新
	common.Corn.SetTask("update", "0 0 16 * * *", func() {
		if c.Workday.TodayIs() {
			logs.PrintErr(c.UpdateCodes(c.GetStockCodes()))
		}
	})

	//等待客户端退出
	<-c.Pool.Done()
}
