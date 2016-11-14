package main

import (
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strings"
	"syscall"
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

	rcv, err := receiver.New()
	if err != nil {
		log.Fatalln("Set receive serve error ", err)
	}
	rcv.GoServe()

	sender.Start(rcv.Rm)

	http.Start()

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGTERM)
	for sig := range signals {
		if sig == syscall.SIGTERM {
			http.Stop()
			rcv.Stop()
			log.Println("exit SIGTERM", time.Now())
			rcv.Rm.Wait()
			log.Println("exit SIGTERM end", time.Now())
			os.Exit(0)
			//TODO . timeout exit
		} else if sig == syscall.SIGHUP {
			http.Stop()
			rcv.Stop()
			log.Println("exit SIGHUP", time.Now())
			os.Setenv("_GRACEFUL_RESTART", "true")
			execSpec := &syscall.ProcAttr{
				Env:   os.Environ(),
				Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
			}
			// Fork exec the new version of your server
			fork, err := syscall.ForkExec(os.Args[0], os.Args, execSpec)
			if err != nil {
				log.Fatalln("Fail to fork", err)
			}

			log.Println("SIGHUP received: fork-exec to", fork)
			// Wait for all conections to be finished
			rcv.Rm.Wait()
			log.Println(os.Getpid(), "Server gracefully shutdown")
			log.Println("exit SIGHUP", time.Now())
			// Stop the old server, all the connections have been closed and the new one is running
			os.Exit(0)
		}
	}
}
