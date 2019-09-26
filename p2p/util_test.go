package p2p

import (
	"github.com/stretchr/testify/assert"
	"net"
	"strings"
	"testing"
)

func TestGetLocalIp(t *testing.T) {
	assert.NotEqual(t, net.ParseIP("127.0.0.1"), GetLocalIp())
}

func BenchmarkGetLocalIp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		assert.NotEqual(b, net.ParseIP("127.0.0.1"), GetLocalIp())
	}
}

func BenchmarkCheckLocalIP(b *testing.B) {
	ips := []string{"", "127.0.0.1", "localhost", GetLocalIp().String()}
	for i := 0; i < b.N; i++ {
		for _, v := range ips {
			assert.Equal(b, true, CheckLocalIP(v))
		}
	}
}

func TestGetBroadcastIPs(t *testing.T) {
	ips, err := GetBroadcastIPs()
	assert.NoError(t, err)
	splitIP := strings.Split(ips[0].String(), ".")
	assert.Equal(t, 4, len(splitIP))
	assert.Equal(t, "255", splitIP[3])
	t.Log(ips)
}

func BenchmarkGetBroadcastIPs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ips, err := GetBroadcastIPs()
		assert.NoError(b, err)
		assert.Equal(b, 1, len(ips))
		splitIP := strings.Split(ips[0].String(), ".")
		assert.Equal(b, 4, len(splitIP))
		assert.Equal(b, "255", splitIP[3])
	}
}
