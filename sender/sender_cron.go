package sender

import (
	"time"

	pfc "github.com/baishancloud/goperfcounter"
	"github.com/baishancloud/octopux-swtfr/g"
	"github.com/toolkits/container/list"
)

const (
	DefaultProcCronPeriod = time.Duration(5) * time.Second    //ProcCron的周期,默认1s
	DefaultLogCronPeriod  = time.Duration(3600) * time.Second //LogCron的周期,默认300s
)

// send_cron程序入口
func startSenderCron(server *g.ReceiverStatusManager) {
	go startProcCron(server)
	go startLogCron(server)
}

func startProcCron(server *g.ReceiverStatusManager) {
	server.Add(1)
	defer server.Done()
	for {
		time.Sleep(DefaultProcCronPeriod)
		refreshSendingCacheSize()
		if server.IsStop() {
			return
		}
	}
}

func startLogCron(server *g.ReceiverStatusManager) {
	server.Add(1)
	defer server.Done()
	for {
		time.Sleep(DefaultLogCronPeriod)
		logConnPoolsProc()
		if server.IsStop() {
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

func logConnPoolsProc() {

}
