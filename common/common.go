package common

import (
	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv/cfg/v2"
	influx "github.com/injoyai/goutil/database/influxdb"
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/goutil/task"
)

var (
	DB   *xorms.Engine    //mysql,用于存储一些基本信息
	TSDB *influx.Client   //influxdb,用于存储历史数据
	Corn = task.New()     //定时任务,用于定时请求数据等
	Real = maps.NewSafe() //实时数据,实时策略加载到缓存
)

func init() {
	TSDB = influx.NewHTTPClient(&influx.HTTPOption{
		Database: cfg.GetString("tsdb.database"),
		Addr:     cfg.GetString("tsdb.address"),
		Username: cfg.GetString("tsdb.username"),
		Password: cfg.GetString("tsdb.password"),
	})
}
