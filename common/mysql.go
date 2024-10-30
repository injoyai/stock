package common

import (
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/goutil/database/mysql"
	"github.com/injoyai/logs"
)

func init() {
	if cfg.GetString("db.type") == "mysql" {
		var err error
		DB, err = mysql.NewXorm(cfg.GetString("db.dsn"))
		logs.PanicErr(err)
		DB.SetSyncField()
	}
}
