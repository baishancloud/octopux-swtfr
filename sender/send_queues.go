package sender

import (
	"github.com/baishancloud/octopux-swtfr/g"
	nlist "github.com/toolkits/container/list"
)

func initSendQueues() {
	cfg := g.Config()

	if cfg.Influxdb != nil && cfg.Influxdb.Enabled {
		for tnode, _ := range cfg.Influxdb.Cluster {
			Q := nlist.NewSafeListLimited(DefaultSendQueueMaxSize)
			InfluxdbQueues[tnode] = Q
		}
	}

}
