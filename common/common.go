package common

import (
	"github.com/injoyai/base/maps"
	influx "github.com/injoyai/goutil/database/influxdb"
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/goutil/frame/mux"
	"github.com/injoyai/goutil/task"
)

var (
	DB   *xorms.Engine    //mysql,用于存储一些基本信息
	TSDB *influx.Client   //influxdb,用于存储历史数据
	Corn = task.New()     //定时任务,用于定时请求数据等
	Real = maps.NewSafe() //实时数据,实时策略加载到缓存
	HTTP = mux.New()      //http服务,对外/页面提供接口
)
