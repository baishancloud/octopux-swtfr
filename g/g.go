package g

import (
	"log"
	"runtime"
)

// changelog:
// 0.0.1: init project
// 0.0.4: bugfix: set replicas before add node
// 0.0.8: change receiver, mv proc cron to proc pkg, add readme, add gitversion, add config reload, add trace tools
// 0.0.9: fix bugs of conn pool(use transfer's private conn pool, named & minimum)
// 0.0.10: use more efficient proc & sema, rm conn_pool status log
// 0.0.11: fix bug: all graphs' traffic delined when one graph broken down, modify retry interval
// 0.0.14: support sending multi copies to graph node, align ts for judge, add filter
// 0.1.4: 添加influxdb存储支持；用于流量采集系统，修改程序名称以区分；修改程序启功方式支持supervisor。
// 0.1.5: 修改项目名称
// 0.1.7：删除 judge 和 graph 部分
// 0.1.9: 添加 mallard pfc统计
// 0.2.0: 添加优雅重启支持
// 1.0.0: 修改打包方式

const (
	VERSION      = "1.0.0"
	GAUGE        = "GAUGE"
	COUNTER      = "COUNTER"
	DERIVE       = "DERIVE"
	DEFAULT_STEP = 60
	MIN_STEP     = 30
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
