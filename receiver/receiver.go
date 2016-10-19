package receiver

import "github.com/baishancloud/octopux-swtfr/receiver/rpc"

func Start() {
	go rpc.StartRpc()

}
