package sender

import (
	"log"
	"time"

	"github.com/baishancloud/swtfr/g"
	"github.com/baishancloud/swtfr/proc"
	"github.com/influxdata/influxdb/client/v2"
	cmodel "github.com/open-falcon/common/model"
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
	judgeConcurrent := cfg.Judge.MaxIdle
	graphConcurrent := cfg.Graph.MaxIdle
	if influxdbConcurrent < 1 {
		influxdbConcurrent = 1
	}

	if judgeConcurrent < 1 {
		judgeConcurrent = 1
	}

	if graphConcurrent < 1 {
		graphConcurrent = 1
	}

	// init send go-routines
	if cfg.Influxdb.Enabled {
		for node, _ := range cfg.Influxdb.Cluster {
			queue := InfluxdbQueues[node]
			go forward2InfluxdbTask(queue, node, influxdbConcurrent)
		}
	}
	for node, _ := range cfg.Judge.Cluster {
		queue := JudgeQueues[node]
		go forward2JudgeTask(queue, node, judgeConcurrent)
	}

	for node, nitem := range cfg.Graph.Cluster2 {
		for _, addr := range nitem.Addrs {
			queue := GraphQueues[node+addr]
			go forward2GraphTask(queue, node, addr, graphConcurrent)
		}
	}

	if cfg.Graph.Migrating {
		for node, cnodem := range cfg.Graph.ClusterMigrating2 {
			for _, addr := range cnodem.Addrs {
				queue := GraphMigratingQueues[node+addr]
				go forward2GraphMigratingTask(queue, node, addr, graphConcurrent)
			}
		}
	}

}

// Judge定时任务, 将 Judge发送缓存中的数据 通过rpc连接池 发送到Judge
func forward2JudgeTask(Q *list.SafeListLimited, node string, concurrent int) {
	batch := g.Config().Judge.Batch // 一次发送,最多batch条数据
	addr := g.Config().Judge.Cluster[node]
	sema := nsema.NewSemaphore(concurrent)

	for {
		items := Q.PopBackBy(batch)
		count := len(items)
		if count == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}

		judgeItems := make([]*cmodel.JudgeItem, count)
		for i := 0; i < count; i++ {
			judgeItems[i] = items[i].(*cmodel.JudgeItem)
		}

		//	同步Call + 有限并发 进行发送
		sema.Acquire()
		go func(addr string, judgeItems []*cmodel.JudgeItem, count int) {
			defer sema.Release()

			resp := &cmodel.SimpleRpcResponse{}
			var err error
			sendOk := false
			for i := 0; i < 3; i++ { //最多重试3次
				err = JudgeConnPools.Call(addr, "Judge.Send", judgeItems, resp)
				if err == nil {
					sendOk = true
					break
				}
				time.Sleep(time.Millisecond * 10)
			}

			// statistics
			if !sendOk {
				log.Printf("send judge %s:%s fail: %v", node, addr, err)
				proc.SendToJudgeFailCnt.IncrBy(int64(count))
			} else {
				proc.SendToJudgeCnt.IncrBy(int64(count))
			}
		}(addr, judgeItems, count)
	}
}

// Graph定时任务, 将 Graph发送缓存中的数据 通过rpc连接池 发送到Graph
func forward2GraphTask(Q *list.SafeListLimited, node string, addr string, concurrent int) {
	batch := g.Config().Graph.Batch // 一次发送,最多batch条数据
	sema := nsema.NewSemaphore(concurrent)

	for {
		items := Q.PopBackBy(batch)
		count := len(items)
		if count == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}

		graphItems := make([]*cmodel.GraphItem, count)
		for i := 0; i < count; i++ {
			graphItems[i] = items[i].(*cmodel.GraphItem)
		}

		sema.Acquire()
		go func(addr string, graphItems []*cmodel.GraphItem, count int) {
			defer sema.Release()

			resp := &cmodel.SimpleRpcResponse{}
			var err error
			sendOk := false
			for i := 0; i < 3; i++ { //最多重试3次
				err = GraphConnPools.Call(addr, "Graph.Send", graphItems, resp)
				if err == nil {
					sendOk = true
					break
				}
				time.Sleep(time.Millisecond * 10)
			}

			// statistics
			if !sendOk {
				log.Printf("send to graph %s:%s fail: %v", node, addr, err)
				proc.SendToGraphFailCnt.IncrBy(int64(count))
			} else {
				proc.SendToGraphCnt.IncrBy(int64(count))
			}
		}(addr, graphItems, count)
	}
}

// Graph定时任务, 进行数据迁移时的 数据冗余发送
func forward2GraphMigratingTask(Q *list.SafeListLimited, node string, addr string, concurrent int) {
	batch := g.Config().Graph.Batch // 一次发送,最多batch条数据
	sema := nsema.NewSemaphore(concurrent)

	for {
		items := Q.PopBackBy(batch)
		count := len(items)
		if count == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}

		graphItems := make([]*cmodel.GraphItem, count)
		for i := 0; i < count; i++ {
			graphItems[i] = items[i].(*cmodel.GraphItem)
		}

		sema.Acquire()
		go func(addr string, graphItems []*cmodel.GraphItem, count int) {
			defer sema.Release()

			resp := &cmodel.SimpleRpcResponse{}
			var err error
			sendOk := false
			for i := 0; i < 3; i++ { //最多重试3次
				err = GraphMigratingConnPools.Call(addr, "Graph.Send", graphItems, resp)
				if err == nil {
					sendOk = true
					break
				}
				time.Sleep(time.Millisecond * 10) //发送失败了,睡10ms
			}

			// statistics
			if !sendOk {
				log.Printf("send to graph migrating %s:%s fail: %v", node, addr, err)
				proc.SendToGraphMigratingFailCnt.IncrBy(int64(count))
			} else {
				proc.SendToGraphMigratingCnt.IncrBy(int64(count))
			}
		}(addr, graphItems, count)
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

			for i := 0; i < retry; i++ { //最多重试3次
				err = InfluxdbConnPools.Send(addr, pts)
				if err == nil {
					proc.SendToInfluxdbCnt.IncrBy(int64(len(pts)))
					break
				}
				time.Sleep(time.Millisecond * 10)
			}

			// statistics
			if err != nil {
				log.Printf("send to tsdb %s:%s fail: %v", node, addr, err)
				proc.SendToInfluxdbFailCnt.IncrBy(int64(len(pts)))
				return
			}
		}(addr, pts)
	}
}
