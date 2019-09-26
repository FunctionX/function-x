package p2p

type Handler interface {
	Handler(c *Context)
}

type EventHandlerFunc func(c *Context)

func (f EventHandlerFunc) Handler(c *Context) {
	f(c)
}

type EventHandler struct {
	messages   map[string]Message
	evHandlers map[Command][]EventHandlerFunc
}

func NewEventHandler(messages map[string]Message) *EventHandler {
	handler := &EventHandler{}
	if messages == nil || len(messages) == 0 {
		handler.messages = map[string]Message{string(MsgMagic[:]): &Msg{}, string(MsgMagic[:]): &Msg{}}
	} else {
		handler.messages = messages
	}
	handler.evHandlers = make(map[Command][]EventHandlerFunc)
	return handler
}

func (e *EventHandler) DoSomething(c *Context) {
	logger.Debug("Handler DoSomething", "addr", c.IP, "command", c.command, "event", EventInfoKV[c.command], "NodeName", c.NodeName)
	if handlers, ok := e.evHandlers[c.command]; ok {
		for _, handler := range handlers {
			go handler.Handler(c)
		}
	}
}

func (e *EventHandler) RegisterEventHandler(command Command, handler ...EventHandlerFunc) {
	e.evHandlers[command] = append(e.evHandlers[command], handler...)
}

func (e *EventHandler) GetMessage(magic []byte) Message {
	if message := e.messages[string(magic)]; message != nil {
		return message.NewMessage()
	}
	return nil
}
