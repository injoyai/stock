package main

import (
	"fmt"
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx"
)

func main() {
	c, err := tdx.Dial("124.71.187.122")
	logs.PanicErr(err)
	resp, err := c.GetKlineDay("sz000001", 0, 800)
	logs.PanicErr(err)
	for _, v := range resp.List {
		fmt.Println(v)
	}
}
