package sender

import (
	"time"

	pfc "github.com/baishancloud/goperfcounter"
	"github.com/baishancloud/octopux-swtfr/g"
	"github.com/toolkits/container/list"
)

const (
	DefaultProcCronPeriod = time.Duration(10) * time.Second //ProcCron的周期,默认1s
)

// send_cron程序入口
func startSenderCron(server *g.ReceiverStatusManager) {
	go startProcCron(server)
}

func startProcCron(server *g.ReceiverStatusManager) {
	server.Add(1)
	defer server.Done()
	for {
		time.Sleep(DefaultProcCronPeriod)
		refreshSendingCacheSize()
		if server.IsRun() == false {
			return
		}
	}
}

func refreshSendingCacheSize() {
	pfc.Gauge("SWTFRInfluxdbQueueSize", calcSendCacheSize(InfluxdbQueues))

}
func calcSendCacheSize(mapList map[string]*list.SafeListLimited) int64 {
	var cnt int64 = 0
	for _, list := range mapList {
		if list != nil {
			cnt += int64(list.Len())
		}
	}
	return cnt
}
