package common

import (
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/logs"
)

func init() {
	if cfg.GetString("db.type") == "sqlite" {
		var err error
		DB, err = sqlite.NewXorm(cfg.GetString("db.dsn"))
		logs.PanicErr(err)
	}
}
