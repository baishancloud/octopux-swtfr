package sender

import (
	"log"
	"time"

	pfc "github.com/baishancloud/goperfcounter"
	"github.com/baishancloud/octopux-swtfr/g"
	"github.com/influxdata/influxdb/client/v2"
	nsema "github.com/toolkits/concurrent/semaphore"
	"github.com/toolkits/container/list"
)

// send
const (
	DefaultSendTaskSleepInterval = time.Millisecond * 50 //默认睡眠间隔为50ms
)

// TODO 添加对发送任务的控制,比如stop等
func startSendTasks() {
	cfg := g.Config()
	// init semaphore
	influxdbConcurrent := cfg.Influxdb.MaxIdle

	if influxdbConcurrent < 1 {
		influxdbConcurrent = 1
	}

	// init send go-routines
	if cfg.Influxdb != nil && cfg.Influxdb.Enabled {
		for node, _ := range cfg.Influxdb.Cluster {
			queue := InfluxdbQueues[node]
			go forward2InfluxdbTask(queue, node, influxdbConcurrent)
		}
	}

}

// Tsdb定时任务, 将 Tsdb发送缓存中的数据 通过api连接池 发送到Tsdb
func forward2InfluxdbTask(Q *list.SafeListLimited, node string, concurrent int) {

	batch := g.Config().Influxdb.Batch // 一次发送,最多batch条数据
	sema := nsema.NewSemaphore(concurrent)
	addr := g.Config().Influxdb.Cluster[node]
	retry := g.Config().Influxdb.MaxRetry

	for {
		items := Q.PopBackBy(batch)
		count := len(items)
		if count == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}
		pts := make([]*client.Point, count)
		for i := 0; i < count; i++ {
			pts[i] = items[i].(*client.Point)
		}

		sema.Acquire()
		go func(addr string, itemList []*client.Point) {
			defer sema.Release()
			var err error
			start := time.Now()

			for i := 0; i < retry; i++ { //最多重试3次
				err = InfluxdbConnPools.Send(addr, pts)
				if err == nil {
					pfc.Meter("SWTFRSendCnt"+node, int64(len(pts)))
					break
				}
				time.Sleep(time.Millisecond * 10)
			}

			// statistics
			if err != nil {
				log.Printf("send to tsdb %s:%s fail: %v", node, addr, err)
				pfc.Meter("SWTFRSendFail"+node, int64(len(pts)))
			}
			pfc.Histogram("SWTFRSendTime"+node, int64(time.Since(start)/time.Millisecond))
		}(addr, pts)
	}
}
