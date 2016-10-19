package sender

import (
	"log"
	"time"

	pfc "github.com/baishancloud/goperfcounter"
	"github.com/baishancloud/octopux-swtfr/g"
	cpool "github.com/baishancloud/octopux-swtfr/sender/conn_pool"
	"github.com/influxdata/influxdb/client/v2"
	cmodel "github.com/open-falcon/common/model"
	nlist "github.com/toolkits/container/list"
)

const (
	DefaultSendQueueMaxSize = 102400 //10.24w
)

// 服务节点的一致性哈希环
// pk -> node
var ()

// 发送缓存队列
// node -> queue_of_data
var (
	InfluxdbQueues = make(map[string]*nlist.SafeListLimited)
)

// 连接池
// node_address -> connection_pool
var (
	InfluxdbConnPools *cpool.InfluxdbConnPoolHelper
)

// 初始化数据发送服务, 在main函数中调用
func Start() {
	initConnPools()
	initSendQueues()
	initNodeRings()
	// SendTasks依赖基础组件的初始化,要最后启动
	startSendTasks()
	startSenderCron()
	log.Println("send.Start, ok")
}

// 将数据 打入所有的Tsdb的发送缓存队列, 相互备份
func Push2TsdbSendQueue(items []*cmodel.MetaData) {
	removeMetrics := g.Config().Influxdb.RemoveMetrics
	//log.Printf("Push2TsdbSendQueue")
	for _, item := range items {
		b, ok := removeMetrics[item.Metric]
		//log.Printf ("select:%V,%V,%V", b, ok,item )
		if b && ok {
			continue
		}
		influxPoint := Convert2InfluxPoint(item)
		errCnt := 0
		for _, Q := range InfluxdbQueues {
			if !Q.PushFront(influxPoint) {
				errCnt += 1
			}
		}

		// statistics
		if errCnt > 0 {
			pfc.Meter("SendToInfluxdbDropCnt", int64(errCnt))
		}

	}
}

// 转化为tsdb格式
func Convert2InfluxPoint(d *cmodel.MetaData) *client.Point {
	d.Tags["Endpoint"] = d.Endpoint
	pt, _ := client.NewPoint(
		d.Metric,
		d.Tags,
		map[string]interface{}{"value": d.Value},
		time.Unix(d.Timestamp, 0),
	)
	return pt
}

func alignTs(ts int64, period int64) int64 {
	return ts - ts%period
}
