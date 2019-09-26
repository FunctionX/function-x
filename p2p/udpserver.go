package p2p

import (
	"fmt"
	"net"
	"os"
	"time"
)

type UdpServer struct {
	Port          int
	udpConn       *net.UDPConn
	handler       *EventHandler
	BroadcastAddr []*net.UDPAddr
	ServerIP      []net.IP
}

var udpServer *UdpServer

func NewUDPServer(port int, handler *EventHandler) *UdpServer {
	if port == 0 {
		panic("UDP port not empty")
	}
	if handler == nil {
		panic("TCP EventHandler not empty")
	}
	udpServer = &UdpServer{}
	udpServer.Port = port
	udpServer.setBroadcastAdders()
	udpServer.handler = handler
	return udpServer
}

func (s *UdpServer) Start() {
	addr, err := net.ResolveUDPAddr(udp, fmt.Sprintf(":%d", s.Port))
	if err != nil {
		panic(err.Error())
	}
	udpConn, err := net.ListenUDP(udp, addr)
	if err != nil {
		panic(err.Error())
	}
	s.udpConn = udpConn
	go s.broadcast()

	for {
		buffer := make([]byte, udpReceiveLen)
		length, addr, err := s.udpConn.ReadFromUDP(buffer)
		if err != nil {
			logger.Error("======== UDP start read data", "err", err.Error())
			continue
		}
		if CheckLocalIP(addr.IP.String()) {
			//logger.Debug("======== UDP IP is local", "addr", addr.IP)
			continue
		}
		message := s.handler.GetMessage(buffer[:2])
		if message == nil {
			logger.Error("======== UDP magic not exist", "addr", addr.IP, "magic", string(buffer[:2]))
			continue
		}
		if length < message.GetHeadLen() {
			logger.Error("======== UDP start data length", "addr", addr.IP, "len", length)
			continue
		}
		bodyLen, err := message.UnmarshalBinary(buffer)
		if err != nil {
			logger.Error("======== UDP unmarshal message", "addr", addr.IP, "len", length)
			continue
		}
		if bodyLen > 0 && uint32(length) > uint32(message.GetHeadLen())+bodyLen {
			message.SetBody(buffer[message.GetHeadLen():bodyLen])
		}
		message.Log(addr.IP, "UDP receive msg <<<<<")
		if message.GetCommand() == CommandNodeDiscovery && os.Getenv(NodeType) == Server {
			go tcpServer.NewTCPConn(addr.IP)
			continue
		}
		if message.GetCommand() == CommandServer {
			s.SetServerIP(addr.IP)
		}
		//message.Handler(addr.IP)
	}
}

func (s *UdpServer) broadcast() {
	logger.Info("======== UDP start broadcast", "broadcastAddr", s.BroadcastAddr)
	defer func() {
		panic("======== UDP Broadcast task aborted")
	}()
	ticker := time.NewTicker(udpTimer * time.Second)
	for range ticker.C {
		for _, addr := range s.BroadcastAddr {
			s.WriteToUDP(NewMsg(CommandNodeDiscovery, nil), addr)
		}
	}
}

func (s *UdpServer) WriteToUDP(message Message, addr *net.UDPAddr) (err error) {
	data, err := message.MarshalBinary()
	if err != nil {
		logger.Error("======== UDP WriteToUDP MarshalBinary", "err", err.Error())
		return err
	}
	_, err = s.udpConn.WriteToUDP(data, addr)
	if err != nil {
		logger.Error("======== UDP WriteToUDP", "err", err.Error())
		s.setBroadcastAdders()
		return err
	}
	message.Log(addr.IP, "UDP send msg ===============>")
	return
}

func (s *UdpServer) setBroadcastAdders() (adders []*net.UDPAddr) {
	ips, err := GetBroadcastIPs()
	if err != nil {
		panic(err.Error())
	}
	for _, ip := range ips {
		addr, err := net.ResolveUDPAddr(udp, fmt.Sprintf("%s:%d", ip.String(), s.Port))
		if err != nil {
			panic(err.Error())
		}
		adders = append(adders, addr)
	}
	s.BroadcastAddr = adders
	return
}

func GetServerIP() []net.IP {
	return udpServer.ServerIP
}

func (s *UdpServer) SetServerIP(ip net.IP) {
	if len(s.ServerIP) == 0 {
		s.ServerIP = append(s.ServerIP, ip)
		return
	}
	for i, v := range s.ServerIP {
		if v.Equal(ip) {
			s.ServerIP = append(s.ServerIP[:i], s.ServerIP[i+1:]...)
			break
		}
	}
	var tmp = s.ServerIP
	s.ServerIP = append([]net.IP{}, ip)
	s.ServerIP = append(s.ServerIP, tmp...)
}
