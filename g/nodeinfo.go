package g

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/toolkits/file"
)

type Port struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type NodeID struct {
	ID string `json:"id"`
}

type NodeInfo struct {
	Manage_ip     string `json:"manage_ip"`
	Node          NodeID `json:"node"`
	Id            string `json:"id"`
	Nodeout_ports []Port `json:"nodeout_ports"`
}

type Nodefile struct {
	errno int        `json:"errno"`
	Data  []NodeInfo `json:"data"`
}

var (
	NodeFilePath string
	nodemap      *map[string]string
	lock         = new(sync.RWMutex)
)

func NodeMap() *map[string]string {
	lock.RLock()
	defer lock.RUnlock()
	return nodemap
}

func ParseNodeConfig(cfg string) {
	if cfg == "" {
		log.Println("请配置nodepath")
		cfg = "node.file"
	}

	if !file.IsExist(cfg) {
		log.Println("node file:", cfg, "is not existent.`")
		return
	}

	NodeFilePath = cfg

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Println("read node file:", cfg, "fail:", err)
		return
	}

	c := Nodefile{}
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Println("parse node file:", cfg, "fail:", err)
		return
	}

	nm := make(map[string]string)
	for _, node := range c.Data {
		if len(node.Nodeout_ports) == 0 {
			continue
		}
		//fmt.Printf("%v\n", node)
		//fmt.Printf("%#v\n", len(node.Nodeout_ports))
		for _, port := range node.Nodeout_ports {
			nm[node.Manage_ip+port.Name] = node.Node.ID
			//fmt.Printf("%s%s:%s\n", node.Manage_ip, port.Name, node.Node.ID)
		}
	}
	lock.Lock()
	defer lock.Unlock()

	nodemap = &nm

	log.Println("read node file:", cfg, "successfully")
}
func cronNodemap() {
	for {
		time.Sleep(600 * time.Second)
		ParseNodeConfig(Config().NodePath)
	}
}
func InitNodemap() {
	ParseNodeConfig(Config().NodePath)
	go cronNodemap()
	//fmt.Printf("%#v\n", nodemap)
}
