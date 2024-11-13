package api

import (
	"github.com/injoyai/goutil/frame/mux"
	"github.com/injoyai/stock/common"
)

func Run(port int) error {

	common.HTTP.ALL("/api/minute/kline/ws", func(r *mux.Request) {
		code := r.GetString("code")
		_ = code
	})

	return common.HTTP.SetPort(port).Run()
}
