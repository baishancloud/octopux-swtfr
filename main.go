package main

import (
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	_ "github.com/baishancloud/goperfcounter"
	"github.com/baishancloud/octopux-swtfr/g"
	"github.com/baishancloud/octopux-swtfr/http"
	"github.com/baishancloud/octopux-swtfr/receiver"
	"github.com/baishancloud/octopux-swtfr/sender"
)

var (
	pid      int
	progname string
)

func init() {
	pid = os.Getpid()
	paths := strings.Split(os.Args[0], "/")
	paths = strings.Split(paths[len(paths)-1], string(os.PathSeparator))
	progname = paths[len(paths)-1]
	runtime.MemProfileRate = 1
}
func saveHeapProfile() {
	runtime.GC()
	f, err := os.Create(fmt.Sprintf("prof/heap_%s_%d_%s.prof", progname, pid, time.Now().Format("2006_01_02_03_04_05")))
	if err != nil {
		return
	}
	defer f.Close()
	pprof.Lookup("heap").WriteTo(f, 1)
}
func main() {
	//defer saveHeapProfile()
	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")
	versionGit := flag.Bool("vg", false, "show version")
	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}
	if *versionGit {
		fmt.Println(g.VERSION, g.COMMIT)
		os.Exit(0)
	}

	// global config
	g.ParseConfig(*cfg)

	sender.Start()
	receiver.Start()

	//go func() {
	//	nethttp.ListenAndServe(":6789", nil)
	//}()

	// http
	http.Start()

	select {}
}
