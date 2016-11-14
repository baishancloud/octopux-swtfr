package sender

import (
	"github.com/baishancloud/octopux-swtfr/g"
	cpool "github.com/baishancloud/octopux-swtfr/sender/connpool"
	nset "github.com/toolkits/container/set"
)

func initConnPools() {
	cfg := g.Config()

	if cfg.Influxdb != nil && cfg.Influxdb.Enabled {
		influxdbInstances := nset.NewStringSet()
		for _, instance := range cfg.Influxdb.Cluster {
			influxdbInstances.Add(instance)
		}
		InfluxdbConnPools = cpool.CreateInfluxdbCliPools(cfg.Influxdb.MaxConns, cfg.Influxdb.MaxIdle,
			cfg.Influxdb.ConnTimeout, cfg.Influxdb.CallTimeout, influxdbInstances.ToSlice())

	}

}

func DestroyConnPools() {

	InfluxdbConnPools.Destroy()
}
