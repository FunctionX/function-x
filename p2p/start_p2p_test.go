package p2p

import (
	"runtime"
	"testing"
)

func TestStartP2P(t *testing.T) {
	//if testing.Verbose() {
	//	t.Skip("start http server")
	//}
	StartP2PServer(NewEventHandler(nil))
}

func TestOS(t *testing.T) {
	t.Log(runtime.GOOS)
}
