package receiver

import (
	"github.com/baishancloud/octopux-swtfr/receiver/rpc"
	"github.com/baishancloud/octopux-swtfr/receiver/socket"
)

func Start() {
	go rpc.StartRpc()
	go socket.StartSocket()
}
