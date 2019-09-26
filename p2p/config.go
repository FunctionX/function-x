package p2p

type Command int16

const (
	CloseUDP     Command = 0
	HeartbeatUDP Command = 1

	// UDP
	CommandNodeDiscovery Command = 1
	CommandServer Command = 2
	CommandNodeCredit Command = 5

	// TCP
	CommandHeartbeat Command = 3
	CommandHeartbeatResponse Command = 4
	CommandLongitude Command = 15
	//CommandNodeType Command = 17
	//CommandNodeTypeResp Command = 18

	NodeDiscoveryHandler = 50
	NodeRemoveHandler Command = 51
)

var MsgInfoKV = map[Command]string{
	CommandNodeDiscovery:     "Discovery",
	CommandNodeCredit:        "Credit",
	CommandHeartbeat:         "Heartbeat",
	CommandHeartbeatResponse: "HeartbeatResponse",
	CommandLongitude:         "Longitude",
	CommandServer:            "Server",
}

var EventInfoKV = map[Command]string{
	NodeDiscoveryHandler: "online",
	NodeRemoveHandler:    "offline",
}

const (
	DebugPort         = 30000
	ReleasePort       = 10000
	HeadLen           = 12
	tcpCreateConnTime = 5
	tcpHeartbeatTime  = 4
	tcpTimer          = 2
	udpTimer          = 2
	reconnectWaitTime = 5
	udpReceiveLen     = 64
	NodeClient        = 1
	NodeServer        = 2

	tcp = "tcp4"
	udp = "udp4"

	NodeType = "node"
	Client   = "client"
	Server   = "server"
)

var NodeTagMap = map[int16]string{
	1: "client",
	2: "server",
}
