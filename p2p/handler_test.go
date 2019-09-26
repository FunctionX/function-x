package p2p

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventHandler_RegisterEventHandler(t *testing.T) {
	//defer func() {
	//	assert.NotEmpty(t, recover())
	//}()
	var command Command = 32767
	handler := NewEventHandler(nil)
	handler.RegisterEventHandler(command, nil)
	handler.RegisterEventHandler(command, nil)
}

// test DoSomething
func TestEventHandler_DoSomething(t *testing.T) {
	t1 := time.Now()
	var command Command = 1024
	var i = 0
	context := &Context{
		Body:    []byte("hello"),
		IP:      net.ParseIP("127.0.0.1"),
		Tag:     NodeServer,
		command: command,
	}
	handler := NewEventHandler(nil)
	handler.RegisterEventHandler(command, func(c *Context) {
		assert.Equal(t, context, c)
		i++
		//t.Log("===============>>>", c)
	})
	for i := 0; i < 100000; i++ {
		handler.DoSomething(context)
	}
	t.Log("time", time.Now().UnixNano()/1e6-t1.UnixNano()/1e6, "ms", i)
	time.Sleep(1 * time.Second)
}

func TestEventHandler_GetMessage(t *testing.T) {
	handler := NewEventHandler(nil)
	message1 := handler.GetMessage([]byte{'X', 'B'})
	assert.NotNil(t, message1)
	message2 := handler.GetMessage([]byte{'X', 'B'})
	assert.NotNil(t, message2)
	//fmt.Println(&message1)
	//fmt.Println(&message2)
	//assert.NotEqual(t, &message1, &message2)
}
