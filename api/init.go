package api

import "github.com/injoyai/stock/common"

func Run() error {
	return common.HTTP.Run()
}
