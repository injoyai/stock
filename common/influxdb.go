package common

import (
	"github.com/injoyai/conv/cfg/v2"
	influx "github.com/injoyai/goutil/database/influxdb"
)

func init() {
	if cfg.GetString("tsdb.type") == "influxdb" {
		TSDB = influx.NewHTTPClient(&influx.HTTPOption{
			Database: cfg.GetString("tsdb.database"),
			Addr:     cfg.GetString("tsdb.address"),
			Username: cfg.GetString("tsdb.username"),
			Password: cfg.GetString("tsdb.password"),
		})
	}
}
