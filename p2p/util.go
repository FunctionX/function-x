package p2p

import (
	"net"
)

func GetLocalIp() net.IP {
	address, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	for _, addr := range address {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP
			}
		}
	}
	return nil
}

func CheckLocalIP(ip string) bool {
	if IP := GetLocalIp(); IP == nil {
		return true
	} else {
		return ip == IP.String() || ip == "127.0.0.1" || ip == "" || ip == "localhost"
	}
}

func GetBroadcastIPs() (ips []net.IP, err error) {
	inetAddresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range inetAddresses {
		if IPNet, ok := addr.(*net.IPNet); ok && !IPNet.IP.IsLoopback() {
			if ip4 := IPNet.IP.To4(); ip4 != nil {
				ip := net.IP(make([]byte, 4))
				for i := range ip4 {
					ip[i] = ip4[i] | ^IPNet.Mask[i]
				}
				ips = append(ips, ip)
			}
		}
	}
	return ips, nil
}

func compressByte(data []byte) (res []byte) {
	for i := 0; i < len(data); i++ {
		if data[i] != '\n' && data[i] != '\t' && data[i] != ' ' {
			res = append(res, data[i])
		}
	}
	return
}
