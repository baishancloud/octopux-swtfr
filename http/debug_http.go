package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/baishancloud/swtfr/sender"
)

func configDebugHttpRoutes() {
	// conn pools
	http.HandleFunc("/debug/connpool/", func(w http.ResponseWriter, r *http.Request) {
		urlParam := r.URL.Path[len("/debug/connpool/"):]
		args := strings.Split(urlParam, "/")

		argsLen := len(args)
		if argsLen < 1 {
			w.Write([]byte(fmt.Sprintf("bad args\n")))
			return
		}

		var result string
		receiver := args[0]
		switch receiver {
		case "tsdb":
			result = strings.Join(sender.InfluxdbConnPools.Proc(), "\n")
		case "judge":
			result = strings.Join(sender.JudgeConnPools.Proc(), "\n")
		case "graph":
			result = strings.Join(sender.GraphConnPools.Proc(), "\n")
		case "graphmigrating":
			result = strings.Join(sender.GraphMigratingConnPools.Proc(), "\n")
		default:
			result = fmt.Sprintf("bad args, module not exist\n")
		}
		w.Write([]byte(result))
	})
}
