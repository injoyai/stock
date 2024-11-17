package main

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/common"
	"github.com/injoyai/stock/data"
	"github.com/injoyai/stock/data/tdx"
)

func main() {

	//连接客户端
	c, err := tdx.Dial(tdx.Hosts, 10)
	logs.PanicErr(err)

	codes := []string{"sz000005"}

	//更新数据
	logs.PrintErr(c.UpdateCodes(codes, false))

	//每天下午16点进行数据更新
	common.Corn.SetTask("update", "0 0 16 * * *", func() {
		isHoliday, err := data.TodayIsHoliday()
		logs.PanicErr(err)
		logs.PrintErr(c.UpdateCodes(c.GetStockCodes(), isHoliday))
	})

	//等待客户端退出
	<-c.Pool.Done()
}
