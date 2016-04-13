package sender

import (
	"github.com/baishancloud/octopux-swtfr/g"
	nlist "github.com/toolkits/container/list"
)

func initSendQueues() {
	cfg := g.Config()
	for node, _ := range cfg.Judge.Cluster {
		Q := nlist.NewSafeListLimited(DefaultSendQueueMaxSize)
		JudgeQueues[node] = Q
	}
	if cfg.Influxdb.Enabled {
		for tnode, _ := range cfg.Influxdb.Cluster {
			Q := nlist.NewSafeListLimited(DefaultSendQueueMaxSize)
			InfluxdbQueues[tnode] = Q
		}
	}

	for node, nitem := range cfg.Graph.Cluster2 {
		for _, addr := range nitem.Addrs {
			Q := nlist.NewSafeListLimited(DefaultSendQueueMaxSize)
			GraphQueues[node+addr] = Q
		}
	}

	if cfg.Graph.Migrating && cfg.Graph.ClusterMigrating != nil {
		for node, cnode := range cfg.Graph.ClusterMigrating2 {
			for _, addr := range cnode.Addrs {
				Q := nlist.NewSafeListLimited(DefaultSendQueueMaxSize)
				GraphMigratingQueues[node+addr] = Q
			}
		}
	}
}
