package p2p

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net"
	"strings"
	"testing"
	"time"
)

func TestNewUDPServer(t *testing.T) {
	port := 8848
	server := NewUDPServer(port, NewEventHandler(nil))
	assert.NotNil(t, server)
	assert.Equal(t, 1, len(server.BroadcastAddr))
	splitIP := strings.Split(server.BroadcastAddr[0].IP.String(), ".")
	assert.Equal(t, 4, len(splitIP))
	assert.Equal(t, "255", splitIP[3])
	//assert.Equal(t, 2, len(server.BroadcastData))
	assert.Equal(t, port, server.Port)
	assert.Nil(t, server.udpConn)
}

func TestSimulationBroadcastUDP(t *testing.T) {
	s := NewUDPServer(8848, NewEventHandler(nil))
	addr, err := net.ResolveUDPAddr(udp, fmt.Sprintf(":%d", s.Port))
	assert.NoError(t, err)
	udpConn, err := net.ListenUDP(udp, addr)
	assert.NoError(t, err)
	s.udpConn = udpConn
	ticker := time.NewTicker(udpTimer * time.Second)
	i := 0
	for range ticker.C {
		i++
		if i > 2 {
			return
		}
		for _, addr := range s.BroadcastAddr {
			_ = addr
			//for _, msg := range s.BroadcastData {
			//	assert.NoError(t, s.WriteToUDP(msg, addr))
			//}
		}
	}
}

func TestReadMsg(t *testing.T) {
	s := NewUDPServer(8849, NewEventHandler(nil))
	addr, err := net.ResolveUDPAddr(udp, fmt.Sprintf(":%d", s.Port))
	assert.NoError(t, err)
	udpConn, err := net.ListenUDP(udp, addr)
	assert.NoError(t, err)
	s.udpConn = udpConn
	go s.broadcast()
	time.Sleep(1 * time.Second)

	buffer := make([]byte, udpReceiveLen)
	length, addr, err := s.udpConn.ReadFromUDP(buffer)
	assert.NoError(t, err)
	assert.Equal(t, true, length > 10)

	message := s.handler.GetMessage(buffer[:2])
	assert.NotNil(t, message)
	message.UnmarshalBinary(buffer)
	//assert.NoError(t, )

	//message.UdpHandler(addr.IP)
	t.Log(addr.IP)
}

func TestTimer(t *testing.T) {
	if testing.Verbose() {
		t.Skip("start http server")
	}
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		fmt.Println(time.Now().Second())
		if time.Now().Second()%3 == 0 {
			fmt.Println("==========", time.Now().Second())
		}
	}
}

func TestUdpServer_SetServerIP(t *testing.T) {
	server := NewUDPServer(8888, NewEventHandler(nil))

	server.SetServerIP(net.ParseIP("192.168.0.1"))
	fmt.Println(server.ServerIP)

	server.SetServerIP(net.ParseIP("192.168.0.2"))
	fmt.Println(server.ServerIP)

	server.SetServerIP(net.ParseIP("192.168.0.3"))
	fmt.Println(server.ServerIP)

	server.SetServerIP(net.ParseIP("192.168.0.4"))
	fmt.Println(server.ServerIP)

	server.SetServerIP(net.ParseIP("192.168.0.1"))
	fmt.Println(server.ServerIP)

	server.SetServerIP(net.ParseIP("192.168.0.2"))
	fmt.Println(server.ServerIP)

	server.SetServerIP(net.ParseIP("192.168.0.3"))
	fmt.Println(server.ServerIP)

	server.SetServerIP(net.ParseIP("192.168.0.4"))
	fmt.Println(server.ServerIP)
}
