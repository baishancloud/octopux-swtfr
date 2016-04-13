package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/baishancloud/octopux-swtfr/g"
	"github.com/baishancloud/octopux-swtfr/http"
	"github.com/baishancloud/octopux-swtfr/proc"
	"github.com/baishancloud/octopux-swtfr/receiver"
	"github.com/baishancloud/octopux-swtfr/sender"
)

func main() {
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
	g.InitNodemap()
	// proc
	proc.Start()

	sender.Start()
	receiver.Start()

	// http
	http.Start()

	select {}
}
