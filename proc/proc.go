package proc

import (
	nproc "github.com/toolkits/proc"
	"log"
)

// trace
var (
	RecvDataTrace = nproc.NewDataTrace("RecvDataTrace", 3)
)

// filter
var (
	RecvDataFilter = nproc.NewDataFilter("RecvDataFilter", 5)
)

// 统计指标的整体数据
var (
	// 计数统计,正确计数,错误计数, ...
	RecvCnt       = nproc.NewSCounterQps("RecvCnt")
	RpcRecvCnt    = nproc.NewSCounterQps("RpcRecvCnt")
	HttpRecvCnt   = nproc.NewSCounterQps("HttpRecvCnt")
	SocketRecvCnt = nproc.NewSCounterQps("SocketRecvCnt")

	SendToJudgeCnt          = nproc.NewSCounterQps("SendToJudgeCnt")
	SendToInfluxdbCnt       = nproc.NewSCounterQps("SendToTsdbCnt")
	SendToGraphCnt          = nproc.NewSCounterQps("SendToGraphCnt")
	SendToGraphMigratingCnt = nproc.NewSCounterQps("SendToGraphMigratingCnt")

	SendToJudgeDropCnt          = nproc.NewSCounterQps("SendToJudgeDropCnt")
	SendToInfluxdbDropCnt       = nproc.NewSCounterQps("SendToTsdbDropCnt")
	SendToGraphDropCnt          = nproc.NewSCounterQps("SendToGraphDropCnt")
	SendToGraphMigratingDropCnt = nproc.NewSCounterQps("SendToGraphMigratingDropCnt")

	SendToJudgeFailCnt          = nproc.NewSCounterQps("SendToJudgeFailCnt")
	SendToInfluxdbFailCnt       = nproc.NewSCounterQps("SendToTsdbFailCnt")
	SendToGraphFailCnt          = nproc.NewSCounterQps("SendToGraphFailCnt")
	SendToGraphMigratingFailCnt = nproc.NewSCounterQps("SendToGraphMigratingFailCnt")

	// 发送缓存大小
	JudgeQueuesCnt          = nproc.NewSCounterBase("JudgeSendCacheCnt")
	InfluxdbQueuesCnt       = nproc.NewSCounterBase("TsdbSendCacheCnt")
	GraphQueuesCnt          = nproc.NewSCounterBase("GraphSendCacheCnt")
	GraphMigratingQueuesCnt = nproc.NewSCounterBase("GraphMigratingCacheCnt")
)

func Start() {
	log.Println("proc.Start, ok")
}

func GetAll() []interface{} {
	ret := make([]interface{}, 0)

	// recv cnt
	ret = append(ret, RecvCnt.Get())
	ret = append(ret, RpcRecvCnt.Get())
	ret = append(ret, HttpRecvCnt.Get())
	ret = append(ret, SocketRecvCnt.Get())

	// send cnt
	ret = append(ret, SendToJudgeCnt.Get())
	ret = append(ret, SendToInfluxdbCnt.Get())
	ret = append(ret, SendToGraphCnt.Get())
	ret = append(ret, SendToGraphMigratingCnt.Get())

	// drop cnt
	ret = append(ret, SendToJudgeDropCnt.Get())
	ret = append(ret, SendToInfluxdbDropCnt.Get())
	ret = append(ret, SendToGraphDropCnt.Get())
	ret = append(ret, SendToGraphMigratingDropCnt.Get())

	// send fail cnt
	ret = append(ret, SendToJudgeFailCnt.Get())
	ret = append(ret, SendToInfluxdbFailCnt.Get())
	ret = append(ret, SendToGraphFailCnt.Get())
	ret = append(ret, SendToGraphMigratingFailCnt.Get())

	// cache cnt
	ret = append(ret, JudgeQueuesCnt.Get())
	ret = append(ret, InfluxdbQueuesCnt.Get())
	ret = append(ret, GraphQueuesCnt.Get())
	ret = append(ret, GraphMigratingQueuesCnt.Get())

	return ret
}
