package p2p

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"encoding/json"
	"net"
	"os"
	"sync"
)

type Message interface {
	encoding.BinaryMarshaler
	UnmarshalBinary(data []byte) (bodyLen uint32, err error)
	Handler(IP net.IP)
	SetBody(body []byte)
	GetCommand() (command Command)
	GetHeadLen() (len int)
	NewMessage() Message
	ResponseMessage(command Command, data []byte) Message
	Log(IP net.IP, info string)
}

var MsgMagic = [2]byte{'X', 'B'}

type Msg struct {
	Head Head
	Body []byte
}

type Head struct {
	Magic   [2]byte
	Command Command
	Tag     int16
	MsgId   int16
	Len     uint32
}

var msgId int16 = 0
var newMsgMu = sync.Mutex{}

func NewMsg(command Command, data []byte) (msg *Msg) {
	newMsgMu.Lock()
	defer newMsgMu.Unlock()
	msgId += 1
	data = compressByte(data)
	var tag int16 = NodeServer
	if os.Getenv(NodeType) == Client {
		tag = NodeClient
	}
	msg = &Msg{
		Head: Head{
			Magic:   MsgMagic,
			Command: command,
			Tag:     tag,
			MsgId:   msgId,
			Len:     uint32(len(data)),
		},
		Body: data,
	}
	return
}

func (msg *Msg) NewMessage() Message {
	return new(Msg)
}

func (msg *Msg) ResponseMessage(command Command, data []byte) Message {
	var tag int16 = NodeServer
	if os.Getenv(NodeType) == Client {
		tag = NodeClient
	}
	return &Msg{
		Head: Head{
			Magic:   MsgMagic,
			Command: command,
			Tag:     tag,
			MsgId:   msg.Head.MsgId,
			Len:     uint32(len(data)),
		},
		Body: data,
	}
}

func (msg *Msg) GetHeadLen() (len int) {
	return HeadLen
}

func (msg *Msg) GetCommand() (command Command) {
	return msg.Head.Command
}

func (msg *Msg) SetBody(body []byte) {
	msg.Body = body
}

func (msg *Msg) MarshalBinary() (data []byte, err error) {
	buf := bytes.NewBuffer(data)
	for _, field := range []interface{}{msg.Head.Magic, msg.Head.Command, msg.Head.Tag, msg.Head.MsgId, msg.Head.Len, msg.Body} {
		err := binary.Write(buf, binary.BigEndian, field)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (msg *Msg) UnmarshalBinary(headBt []byte) (bodyLen uint32, err error) {
	msg.Head = Head{}
	buf := bytes.NewBuffer(headBt)
	for _, bs := range []interface{}{&msg.Head.Magic, &msg.Head.Command, &msg.Head.Tag, &msg.Head.MsgId, &msg.Head.Len} {
		if err := binary.Read(buf, binary.BigEndian, bs); err != nil {
			return bodyLen, err
		}
	}
	return msg.Head.Len, nil
}

func (msg *Msg) Handler(IP net.IP) {
	var data struct {
		NodeName string `json:"nodeName"`
	}
	if msg.Head.Command == CommandHeartbeatResponse || msg.Head.Command == CommandHeartbeat {
		if len(msg.Body) <= 0 || msg.Head.Len <= 0 {
			return
		}
		if err := json.Unmarshal(msg.Body, &data); err != nil {
			logger.Error("msg handler json unmarshal", "addr", IP, "err", err.Error())
			return
		}
		msg.Head.Command = NodeDiscoveryHandler
	}

	// handler
	context := &Context{NodeName: data.NodeName, IP: IP, Tag: msg.Head.Tag, command: msg.Head.Command, Body: msg.Body}
	tcpServer.handler.DoSomething(context)
}

func (msg *Msg) Log(IP net.IP, info string) {
	logger.Debug(info, "addr", IP, "msg", MsgInfoKV[msg.Head.Command], "msgId", msg.Head.MsgId, "tag", NodeTagMap[msg.Head.Tag], "len", msg.Head.Len, "body", string(msg.Body))
}
