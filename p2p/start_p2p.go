package p2p

import (
	"os"
	"runtime"
)

func StartP2PServer(handler *EventHandler) {
	udpPort := ReleasePort
	//udpPort := DebugPort
	//logger.InitLogger(logger.LvlDebug, "./tmp/logs/udp-server.log")
	if runtime.GOOS == "android" {
		udpPort = ReleasePort
	}
	if err := os.Setenv(NodeType, Server); err != nil {
		panic(err.Error())
	}
	logger.Info("p2p run ......", "port", udpPort, "os", runtime.GOOS, "ip", GetLocalIp())
	go NewUDPServer(udpPort, handler).Start()

	NewTCPServer(udpPort+1, handler).Start()
}
