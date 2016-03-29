package sender

import (
	"github.com/baishancloud/swtfr/g"
	cpool "github.com/baishancloud/swtfr/sender/conn_pool"
	nset "github.com/toolkits/container/set"
)

func initConnPools() {
	cfg := g.Config()

       if cfg.Influxdb.Enabled {
               influxdbInstances := nset.NewStringSet()
               for _, instance := range cfg.Influxdb.Cluster {
                       influxdbInstances.Add(instance)
               }
               InfluxdbConnPools = cpool.CreateInfluxdbCliPools(cfg.Influxdb.MaxConns, cfg.Influxdb.MaxIdle,
                       cfg.Influxdb.ConnTimeout, cfg.Influxdb.CallTimeout, influxdbInstances.ToSlice())

       }
	judgeInstances := nset.NewStringSet()
	for _, instance := range cfg.Judge.Cluster {
		judgeInstances.Add(instance)
	}
	JudgeConnPools = cpool.CreateSafeRpcConnPools(cfg.Judge.MaxConns, cfg.Judge.MaxIdle,
		cfg.Judge.ConnTimeout, cfg.Judge.CallTimeout, judgeInstances.ToSlice())

	// graph
	graphInstances := nset.NewSafeSet()
	for _, nitem := range cfg.Graph.Cluster2 {
		for _, addr := range nitem.Addrs {
			graphInstances.Add(addr)
		}
	}
	GraphConnPools = cpool.CreateSafeRpcConnPools(cfg.Graph.MaxConns, cfg.Graph.MaxIdle,
		cfg.Graph.ConnTimeout, cfg.Graph.CallTimeout, graphInstances.ToSlice())

	// graph migrating
	if cfg.Graph.Migrating && cfg.Graph.ClusterMigrating != nil {
		graphMigratingInstances := nset.NewSafeSet()
		for _, cnode := range cfg.Graph.ClusterMigrating2 {
			for _, addr := range cnode.Addrs {
				graphMigratingInstances.Add(addr)
			}
		}
		GraphMigratingConnPools = cpool.CreateSafeRpcConnPools(cfg.Graph.MaxConns, cfg.Graph.MaxIdle,
			cfg.Graph.ConnTimeout, cfg.Graph.CallTimeout, graphMigratingInstances.ToSlice())
	}
}

func DestroyConnPools() {

	InfluxdbConnPools.Destroy()
	JudgeConnPools.Destroy()
	GraphConnPools.Destroy()
	GraphMigratingConnPools.Destroy()
}
