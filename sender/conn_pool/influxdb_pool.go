package conn_pool

//Influxdb
import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/baishancloud/swtfr/g"
	"github.com/influxdata/influxdb/client/v2"
)

type InfluxdbClient struct {
	name       string
	Address    string
	Username   string
	Password   string
	Database   string
	UserAgent  string
	Precision  string
	Timeout    time.Duration
	UDPPayload int `toml:"udp_payload"`

	cli client.Client
}

func (this InfluxdbClient) Name() string {
	return this.name
}

func (this InfluxdbClient) Closed() bool {
	return this.cli == nil
}

func (this InfluxdbClient) Close() error {
	if this.cli != nil {
		this.cli.Close()
		this.cli = nil
		return nil
	}
	return nil
}

func (this *InfluxdbClient) Connect() error {

	// Backward-compatability with single Influx URL config files
	// This could eventually be removed in favor of specifying the urls as a list
	if this.Address == "" {
		return fmt.Errorf("Influxdb url is nil.")
	}

	switch {
	case strings.HasPrefix(this.Address, "udp"):
		parsed_url, err := url.Parse(this.Address)
		if err != nil {
			return err
		}

		if this.UDPPayload == 0 {
			this.UDPPayload = client.UDPPayloadSize
		}
		c, err := client.NewUDPClient(client.UDPConfig{
			Addr:        parsed_url.Host,
			PayloadSize: this.UDPPayload,
		})
		if err != nil {
			return err
		}
		this.cli = c
	default:
		// If URL doesn't start with "udp", assume HTTP client
		c, err := client.NewHTTPClient(client.HTTPConfig{
			Addr:      this.Address,
			Username:  this.Username,
			Password:  this.Password,
			UserAgent: this.UserAgent,
			Timeout:   this.Timeout,
		})
		if err != nil {
			return err
		}

		// Create Database if it doesn't exist
		_, e := c.Query(client.Query{
			Command: fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", this.Database),
		})

		if e != nil {
			log.Println("Database creation failed: " + e.Error())
		}

		this.cli = c
	}

	return nil
}

func newInfluxdbConnPool(name string, address string, connTimeout time.Duration, maxConns int, maxIdle int) *ConnPool {
	pool := NewConnPool(name, address, maxConns, maxIdle)
	pool.Username = g.Config().Influxdb.Username
	pool.Password = g.Config().Influxdb.Password
	pool.Database = g.Config().Influxdb.Database
	pool.Precision = "s"

	pool.New = func(connName string) (NConn, error) {
		nconn := InfluxdbClient{
			name:      connName,
			Address:   pool.Address,
			Username:  pool.Username,
			Password:  pool.Password,
			Database:  pool.Database,
			Precision: pool.Precision,
			Timeout:   connTimeout,
		}
		err := nconn.Connect()
		if err != nil {
			return nil, err
		}

		return nconn, nil
	}

	return pool
}

type InfluxdbConnPoolHelper struct {
	sync.RWMutex
	M           map[string]*ConnPool
	MaxConns    int
	MaxIdle     int
	ConnTimeout int
	CallTimeout int
}

func (this *InfluxdbConnPoolHelper) Get(address string) (*ConnPool, bool) {
	this.RLock()
	defer this.RUnlock()
	p, exists := this.M[address]
	return p, exists
}

func (this *InfluxdbConnPoolHelper) Destroy() {
	this.Lock()
	defer this.Unlock()
	addresses := make([]string, 0, len(this.M))
	for address := range this.M {
		addresses = append(addresses, address)
	}

	for _, address := range addresses {
		this.M[address].Destroy()
		delete(this.M, address)
	}
}

func CreateInfluxdbCliPools(maxConns, maxIdle, connTimeout, callTimeout int, cluster []string) *InfluxdbConnPoolHelper {
	tp := &InfluxdbConnPoolHelper{
		M: make(map[string]*ConnPool), MaxConns: maxConns, MaxIdle: maxIdle,
		ConnTimeout: connTimeout, CallTimeout: callTimeout,
	}

	ct := time.Duration(tp.ConnTimeout) * time.Millisecond
	for _, address := range cluster {
		if _, exist := tp.M[address]; exist {
			continue
		}
		tp.M[address] = newInfluxdbConnPool(address, address, ct, maxConns, maxIdle)
	}
	return tp
}

func (this *InfluxdbConnPoolHelper) Proc() []string {
	procs := []string{}
	for _, cp := range this.M {
		procs = append(procs, cp.Proc())
	}
	return procs
}

func (this *InfluxdbConnPoolHelper) Send(addr string, points []*client.Point) (err error) {
	connPool, exists := this.Get(addr)
	if !exists {
		return fmt.Errorf("%s has no connection pool", addr)
	}

	conn, err := connPool.Fetch()
	if err != nil {
		return fmt.Errorf("%s get connection fail: conn %v, err %v. proc: %s", addr, conn, err, connPool.Proc())
	}

	cli := conn.(InfluxdbClient)

	done := make(chan error)
	go func() {
		err = cli.Write(points)
		done <- err
	}()

	select {
	case <-time.After(time.Duration(this.CallTimeout) * time.Millisecond):
		connPool.ForceClose(conn)
		return fmt.Errorf("%s, call timeout", conn.Name())
	case err = <-done:
		if err != nil {
			connPool.ForceClose(conn)
			err = fmt.Errorf("%s, call failed, err %v. proc: %s", conn.Name(), err, connPool.Proc())
		} else {
			connPool.Release(conn)
		}
		return err
	}
}

//influx write
func (this *InfluxdbClient) Write(points []*client.Point) error {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  this.Database,
		Precision: this.Precision,
	})

	for _, point := range points {
		bp.AddPoint(point)
	}

	if e := this.cli.Write(bp); e != nil {
		log.Println("ERROR: " + e.Error())
		return errors.New("Could not write to any InfluxDB server in cluster")
	}

	return nil
}
