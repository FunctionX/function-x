package p2p

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func TestNewTCPServer(t *testing.T) {
	server := NewTCPServer(8890, NewEventHandler(nil))
	assert.NotNil(t, server)
	assert.Equal(t, 8890, server.Port)
	assert.NotNil(t, server.broadcastData)

	assert.NotNil(t, server.nodes)
	assert.NotNil(t, server.handler)
}

func TestAddBroadcastDataTcp(t *testing.T) {
	server := NewTCPServer(8890, NewEventHandler(nil))

	assert.NotNil(t, server.broadcastData)
}

func TestTimeDuration(t *testing.T) {
	var counter time.Duration
	var maxDuration time.Duration = 1<<63 - 1
	//counter = maxDuration + 1
	t.Log(maxDuration)
	t.Log("int16 max ", 1<<16/2-1)
	counter = -30
	t.Log(counter%10 == 0)

}

func TestCreateTcpConn(t *testing.T) {
	if testing.Verbose() {
		t.Skip("start http server")
	}
	tcpAddr, err := net.ResolveTCPAddr(tcp, fmt.Sprintf("%s:%d", "192.168.21.163", 40001))
	assert.NoError(t, err)

	conn, err := net.DialTimeout(tcp, tcpAddr.String(), tcpCreateConnTime*time.Second)
	//conn, err := net.DialTCP(tcp, nil, tcpAddr)
	assert.NoError(t, err)
	t.Log("addr", conn.RemoteAddr(), conn)
}

func TestBroadcastLocation(t *testing.T) {
	s := NewTCPServer(8679, NewEventHandler(nil))

	go s.Start()

	type LongitudeInfo struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	addr, _ := net.ResolveTCPAddr(tcp, fmt.Sprintf(":%d", s.Port))
	s.AddNode(&TcpNode{addr: addr})

	time.Sleep(3 * time.Second)
}

func TestAddBroadcastDataTcp2(t *testing.T) {
	all := []byte(`{
	"latitude": 100,
    "longitude": 200
}`)
	var data []byte
	for i := 0; i < len(all); i++ {
		if all[i] != '\n' && all[i] != '\t' && all[i] != ' ' {
			data = append(data, all[i])
		}
	}
	fmt.Println(data)
	fmt.Println(string(data))
}

func TestParseJson(t *testing.T) {
	data := fmt.Sprintf(`{"nodeName":"%v"}`, "")
	t.Log(data)
	res := struct {
		NodeName string `json:"nodeName"`
	}{}
	json.Unmarshal([]byte(data), &res)
	t.Log(res.NodeName)
}
