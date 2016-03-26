package receiver

import (
	"github.com/baishancloud/swtfr/receiver/rpc"
	"github.com/baishancloud/swtfr/receiver/socket"
)

func Start() {
	go rpc.StartRpc()
	go socket.StartSocket()
}
