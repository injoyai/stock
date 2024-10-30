package common

import (
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/goutil/database/mssql"
	"github.com/injoyai/logs"
)

func init() {
	if cfg.GetString("db.type") == "mssql" {
		var err error
		DB, err = mssql.NewXorm(cfg.GetString("db.dsn"))
		logs.PanicErr(err)
	}
}
