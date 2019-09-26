package p2p

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

type TcpServer struct {
	Port          int
	handler       *EventHandler
	nodeCh        chan *TcpNode
	nodes         map[string]*TcpNode
	broadcastData BroadcastData
	sync.Mutex
}

type TcpNode struct {
	addr     *net.TCPAddr
	conn     *net.TCPConn
	msgId    int16
	isServer bool
	sync.Mutex
	isReturn bool
	lastTime time.Time
	isOnline bool
	isStart  bool
}

var tcpServer *TcpServer

func NewTCPServer(port int, handler *EventHandler) *TcpServer {
	if port == 0 {
		panic("TCP port not empty")
	}
	if handler == nil {
		panic("TCP EventHandler not empty")
	}
	tcpServer = &TcpServer{}
	tcpServer.Port = port
	tcpServer.handler = handler
	tcpServer.nodeCh = make(chan *TcpNode)
	tcpServer.broadcastData = BroadcastData{}
	tcpServer.nodes = map[string]*TcpNode{}
	return tcpServer
}

func newNode(addr *net.TCPAddr, isServer bool) *TcpNode {
	return &TcpNode{addr: addr, isServer: isServer, isOnline: true}
}

func AddBroadcastDataTcp(broadcastData BroadcastData) {
	if tcpServer == nil || os.Getenv(NodeType) == Client {
		return
	}
	tcpServer.Lock()
	tcpServer.broadcastData = broadcastData
	tcpServer.Unlock()
}

func SetClientName(name string) {
	if tcpServer == nil {
		return
	}
	tcpServer.Lock()
	tcpServer.broadcastData = BroadcastData{NodeName: name}
	tcpServer.Unlock()
}

// listen TCP request connect
func (s *TcpServer) Start() {
	addr, err := net.ResolveTCPAddr(tcp, fmt.Sprintf(":%d", s.Port))
	if err != nil {
		panic(err)
	}
	tcpListen, err := net.ListenTCP(tcp, addr)
	if err != nil {
		panic(err)
	}
	go s.onTcpNode()
	go s.broadcast()

	for {
		conn, err := tcpListen.AcceptTCP()
		if err != nil {
			logger.Error("TCP AcceptTCP err", "err", err.Error())
			continue
		}
		isExist, tcpNode := s.AddNode(newNode(conn.RemoteAddr().(*net.TCPAddr), true))
		if isExist {
			logger.Debug("TCP node connected 1", "addr", tcpNode.addr.IP)
			conn.Close()
			continue
		}
		tcpNode.conn = conn
		logger.Info("TCP created connect", "addr", tcpNode.addr.IP, "note", "local node client")
		s.nodeCh <- tcpNode
	}
}

// create TCP request connect
func (s *TcpServer) NewTCPConn(IP net.IP) {
	tcpAddr, err := net.ResolveTCPAddr(tcp, fmt.Sprintf("%s:%d", IP, s.Port))
	if err != nil {
		logger.Error("TCP NewTCPConn", "ResolveTCPAddr", err.Error())
		return
	}
	isExist, tcpNode := s.AddNode(newNode(tcpAddr, false))
	if isExist {
		logger.Debug("TCP node connected 2", "addr", IP)
		return
	}
	conn, err := net.DialTimeout(tcp, tcpAddr.String(), tcpCreateConnTime*time.Second)
	if err != nil {
		logger.Error("TCP NewTCPConn", "DialTimeout", err.Error())
		s.RemoveNode(tcpNode)
		return
	}
	tcpNode.conn = conn.(*net.TCPConn)
	logger.Info("TCP created connect", "addr", IP, "note", "local node server")
	s.nodeCh <- tcpNode
}

func (s *TcpServer) onTcpNode() {
	defer func() {
		panic("TCP Connection management protocol exception stop")
	}()
	for {
		node := <-s.nodeCh
		node.conn.SetNoDelay(true)
		node.conn.SetKeepAlive(true)
		node.conn.SetKeepAlivePeriod((tcpHeartbeatTime + 2) * time.Second)
		logger.Info("TCP node", "addr", node.addr.IP, "node is server", node.isServer, "nodes", len(s.nodes))
		go node.start()
	}
}

// node manager
func (node *TcpNode) start() {
	defer func() {
		tcpServer.RemoveNode(node)
		logger.Info("TCP node close XX", "addr", node.addr.IP)
	}()

	node.Lock()
	node.isStart = true
	node.Unlock()
	if node.isServer == false {
		tcpServer.Lock()
		dataInfo := tcpServer.getBroadcastMsg()
		tcpServer.Unlock()
		node.WriteTo(NewMsg(CommandHeartbeat, dataInfo)) // heart
	}
	for {
		if err := node.conn.SetReadDeadline(time.Now().Add(tcpHeartbeatTime * time.Second)); err != nil {
			logger.Warn("TCP set read deadline", "addr", node.addr.IP, "err", err.Error())
			return
		}
		headBt := make([]byte, HeadLen)
		if _, err := node.conn.Read(headBt); err != nil {
			logger.Warn("TCP receive head info", "addr", node.addr.IP, "err", err.Error())
			return
		}
		message := tcpServer.handler.GetMessage(headBt[:2])
		if message == nil {
			logger.Error("TCP messages is nil", "addr", node.addr.IP, "head", headBt)
			return
		}
		length, err := message.UnmarshalBinary(headBt)
		if err != nil {
			logger.Error("TCP parse request head", "err", err.Error())
			return
		}
		if length > 0 {
			body := make([]byte, length)
			if _, err := node.conn.Read(body); err != nil {
				logger.Error("TCP receive body info", "addr", node.addr.IP, "err", err.Error())
				return
			}
			message.SetBody(body)
		}
		message.Log(node.addr.IP, "TCP receive msg <<<<<")
		if message.GetCommand() == CommandHeartbeat {
			if os.Getenv(NodeType) == Client {
				tcpServer.Lock()
				dataInfoMsg := tcpServer.getClientResponseMsg()
				tcpServer.Unlock()
				node.WriteTo(message.ResponseMessage(CommandHeartbeatResponse, dataInfoMsg))
			} else {
				tcpServer.Lock()
				dataInfoMsg := tcpServer.getBroadcastMsg()
				tcpServer.Unlock()
				node.WriteTo(message.ResponseMessage(CommandHeartbeatResponse, dataInfoMsg))
			}
		}
		if message.GetCommand() == CommandHeartbeatResponse {
			node.Lock()
			node.isReturn = true
			node.Unlock()
		}
		message.Handler(node.addr.IP)
	}
}

// TCP Write node
func (node *TcpNode) WriteTo(message Message) (err error) {
	data, err := message.MarshalBinary()
	if err != nil {
		logger.Error("TCP write marshal msg", "err", err.Error())
		return err
	}
	if node.conn == nil {
		logger.Error("TCP node conn is nil", "node", node)
		tcpServer.RemoveNode(node)
		return errors.New("node conn is nil")
	}
	if err = node.conn.SetWriteDeadline(time.Now().Add(3 * time.Second)); err != nil {
		logger.Error("TCP set write deadline", "err", err.Error())
		return err
	}
	if _, err := node.conn.Write(data); err != nil {
		logger.Error("TCP write to data", "err", err.Error())
		tcpServer.RemoveNode(node)
		return err
	}
	message.Log(node.addr.IP, "TCP send msg ===============>")
	return nil
}

// TCP Write by IP
func (s *TcpServer) WriteToTCP(message Message, ip string) (err error) {
	s.Lock()
	defer s.Unlock()

	if node := s.nodes[ip]; node == nil {
		logger.Warn("TCP WriteToTCP node not exist", "addr", ip)
		return errors.New("node not exist, send msg error")
	} else {
		return node.WriteTo(message)
	}
}

// TCP ticker broadcast
func (s *TcpServer) broadcast() {
	defer func() {
		panic("TCP The scheduled task stops abnormally")
	}()
	ticker := time.NewTicker(tcpTimer * time.Second)
	for range ticker.C {
		s.Lock()
		for _, node := range s.nodes {
			node.Lock()
			is := node.isReturn && node.isOnline && node.isStart
			node.Unlock()
			if is {
				go node.WriteTo(NewMsg(CommandHeartbeat, s.getBroadcastMsg()))
			}
		}
		s.Unlock()
	}
}

func (s *TcpServer) AddNode(node *TcpNode) (bool, *TcpNode) {
	s.Lock()
	defer s.Unlock()
	nd := s.nodes[node.addr.IP.String()]
	if nd == nil {
		logger.Info("TCP  ###", "addr", node.addr.IP)
		s.nodes[node.addr.IP.String()] = node
		return false, node
	}
	nd.Lock()
	if nd.isOnline {
		nd.Unlock()
		return true, nd
	} else {
		nd.addr = node.addr
		//nd.conn = node.conn
		nd.isServer = node.isServer
		nd.lastTime = time.Now()
		nd.isOnline = true
		nd.isReturn = false
		nd.msgId = 0
		nd.Unlock()
		return false, nd
	}
}

func (s *TcpServer) RemoveNode(node *TcpNode) {
	node.Lock()
	if node.conn != nil {
		node.conn.Close()
	} else {
		logger.Error("TCP RemoveNode node conn is nil")
	}
	node.isOnline = false
	node.isStart = false
	node.lastTime = time.Now()
	logger.Info("TCP", "addr", node.addr.IP, "lastTime", node.lastTime)
	node.Unlock()

	go node.SendOffLineEvent()
	//s.handler.DoSomething(&Context{IP: node.addr.IP, command: NodeRemoveHandler})
}

func (node *TcpNode) SendOffLineEvent() {
	time.Sleep(reconnectWaitTime * time.Second)
	node.Lock()
	lastTime := node.lastTime
	node.Unlock()
	if time.Now().Sub(lastTime) > reconnectWaitTime*time.Second {
		logger.Error("123 ", "addr", node.addr, "t1", time.Now(), "t2", lastTime)
		tcpServer.handler.DoSomething(&Context{IP: node.addr.IP, command: NodeRemoveHandler})
		return
	}
	logger.Info("123", "addr", node.addr)
	return
}

func (s *TcpServer) getBroadcastMsg() []byte {
	data := s.broadcastData
	if len(data.PositionByte) > 0 {
		ps := Position{}
		err := json.Unmarshal(data.PositionByte, &ps)
		if err != nil {
			logger.Error("json Unmarshal", "err", err.Error())
		} else {
			data.Position = &ps
		}
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		logger.Error("marshal json err", "err", err.Error())
		return nil
	}
	return bytes
}

type BroadcastData struct {
	PositionByte []byte    `json:"-"`
	Position     *Position `json:"position,omitempty"`
	Credit       int64     `json:"credit"`
	NodeName     string    `json:"nodeName,omitempty"`
}

type Position struct {
	Longitude float64 `json:"longitude,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
}

func (s *TcpServer) getClientResponseMsg() []byte {
	if name := s.broadcastData.NodeName; name != "" {
		return []byte(fmt.Sprintf(`{"nodeName":"%v"}`, name))
	}
	return nil
}
