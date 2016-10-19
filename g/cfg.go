package g

import (
	"encoding/json"
	"log"
	"strings"
	"sync"

	"github.com/toolkits/file"
)

type HttpConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
}

type RpcConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
}

type InfluxdbConfig struct {
	Enabled       bool                    `json:"enabled"`
	Batch         int                     `json:"batch"`
	Username      string                  `json:"username"`
	Password      string                  `json:"password"`
	Database      string                  `json:"database"`
	ConnTimeout   int                     `json:"connTimeout"`
	CallTimeout   int                     `json:"callTimeout"`
	MaxConns      int                     `json:"maxConns"`
	MaxIdle       int                     `json:"maxIdle"`
	MaxRetry      int                     `json:"retry"`
	Cluster       map[string]string       `json:"cluster"`
	RemoveMetrics map[string]bool         `json:"remove"`
	Cluster2      map[string]*ClusterNode `json:"cluster2"`
}

type GlobalConfig struct {
	Debug    bool        `json:"debug"`
	NodePath string      `json:"nodepatch"`
	Http     *HttpConfig `json:"http"`
	Rpc      *RpcConfig  `json:"rpc"`

	Influxdb *InfluxdbConfig `json:"influxdb"`
}

var (
	ConfigFile string
	config     *GlobalConfig
	configLock = new(sync.RWMutex)
)

func Config() *GlobalConfig {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}

func ParseConfig(cfg string) {
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, "is not existent. maybe you need `mv cfg.example.json cfg.json`")
	}

	ConfigFile = cfg

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Fatalln("read config file:", cfg, "fail:", err)
	}

	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Fatalln("parse config file:", cfg, "fail:", err)
	}

	// split cluster config
	c.Influxdb.Cluster2 = formatClusterItems(c.Influxdb.Cluster)

	configLock.Lock()
	defer configLock.Unlock()
	config = &c

	log.Println("g.ParseConfig ok, file ", cfg)
}

// CLUSTER NODE
type ClusterNode struct {
	Addrs []string `json:"addrs"`
}

func NewClusterNode(addrs []string) *ClusterNode {
	return &ClusterNode{addrs}
}

// map["node"]="host1,host2" --> map["node"]=["host1", "host2"]
func formatClusterItems(cluster map[string]string) map[string]*ClusterNode {
	ret := make(map[string]*ClusterNode)
	for node, clusterStr := range cluster {
		items := strings.Split(clusterStr, ",")
		nitems := make([]string, 0)
		for _, item := range items {
			nitems = append(nitems, strings.TrimSpace(item))
		}
		ret[node] = NewClusterNode(nitems)
	}

	return ret
}
