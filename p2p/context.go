package p2p

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

type Context struct {
	NodeName string
	IP       net.IP
	Tag      int16
	Body     []byte
	command  Command
}

func NewContext() *Context {
	return &Context{}
}

func (c *Context) SendMsgTCP(command Command, ip *string, msgInfo interface{}) (err error) {
	msg, err := getSendMsg(command, msgInfo)
	if err != nil {
		return err
	}
	if ip == nil {
		// send all
		for _, node := range tcpServer.nodes {
			node.WriteTo(msg)
		}
	} else {
		// send one peer
		tcpServer.WriteToTCP(msg, *ip)
	}
	return nil
}

func (c *Context) SendMsgUDP(command Command, ip *string, msgInfo interface{}) (err error) {
	msg, err := getSendMsg(command, msgInfo)
	if err != nil {
		return err
	}
	if ip == nil {
		// send all
		for _, addr := range udpServer.BroadcastAddr {
			if err = udpServer.WriteToUDP(msg, addr); err != nil {
				return err
			}
		}
	} else {
		// send one
		udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%d", ip, udpServer.Port))
		if err != nil {
			return fmt.Errorf("ip resolve udp addr err:%s", err.Error())
		}
		return udpServer.WriteToUDP(msg, udpAddr)
	}
	return nil
}

func getSendMsg(command Command, msgInfo interface{}) (msg *Msg, err error) {
	if command < 50 {
		return nil, errors.New("command must be above 50")
	}
	switch data := msgInfo.(type) {
	case []byte:
		msg = NewMsg(command, data)
	case string:
		msg = NewMsg(command, []byte(data))
	case struct{}:
		bt, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("msg marshal err:%s", err.Error())
		}
		msg = NewMsg(command, bt)
	default:
		return nil, errors.New("msgInfo invalid format")
	}
	return msg, nil
}
