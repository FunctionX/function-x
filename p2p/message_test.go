package p2p

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestMsg_MarshalBinary(t *testing.T) {
	msg := NewMsg(CommandHeartbeat, nil)
	data, err := msg.MarshalBinary()
	assert.NoError(t, err)
	assert.Equal(t, data, []byte{'X', 'B', 0, 3, 0, 2, 0, 1, 0, 0, 0, 0})
	t.Log(data)

	creditMsg := NewMsg(CommandNodeCredit, []byte(`{"Credit":7}`))
	msgBt, err := creditMsg.MarshalBinary()
	assert.NoError(t, err)
	assert.Equal(t, msgBt, []byte{88, 66, 0, 5, 0, 2, 0, 2, 0, 0, 0, 12, 123, 34, 67, 114, 101, 100, 105, 116, 34, 58, 55, 125})
	t.Log(msgBt)
}

func TestMsg_MarshalBinary2(t *testing.T) {
	msgId = 0
	msg := NewMsg(CommandHeartbeat, nil)
	data, err := msg.MarshalBinary()
	assert.NoError(t, err)
	assert.Equal(t, data, []byte{'X', 'B', 0, 3, 0, 2, 0, 1, 0, 0, 0, 0})
}

func TestMsg_MarshalBinary3(t *testing.T) {
	msgId = 0
	msg := NewMsg(CommandLongitude, []byte("hello"))
	data, err := msg.MarshalBinary()
	assert.NoError(t, err)
	assert.Equal(t, data, []byte{'X', 'B', 0, 15, 0, 2, 0, 1, 0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o'})
}

func TestMsg_Magic(t *testing.T) {
	msg := NewMsg(CommandHeartbeat, nil)
	assert.Equal(t, "XB", string(MsgMagic[:]))
	assert.Equal(t, "XB", string(msg.Head.Magic[:]))
	t.Log(string(msg.Head.Magic[:]))
}

func BenchmarkNewMsg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var tmp int16
		for k, v := range MsgInfoKV {
			msg := NewMsg(k, []byte(v))
			if msg.Head.MsgId != 0 {
				assert.Equal(b, true, msg.Head.MsgId > tmp)
			}
			tmp = msg.Head.MsgId
		}
	}
}

func TestNewMsg(t *testing.T) {
	msgId = 0
	data := []byte("hello")
	msg := &Msg{
		Head: Head{
			Magic:   [2]byte{'X', 'B'},
			Command: CommandLongitude,
			Tag:     NodeServer,
			MsgId:   1,
			Len:     5,
		},
		Body: data,
	}
	assert.Equal(t, msg, NewMsg(CommandLongitude, data))
}

func TestRecursionStruct(t *testing.T) {
	data := []byte("hello")
	msg := &Msg{
		Head: Head{
			Magic:   [2]byte{'X', 'B'},
			Command: 1000,
			Tag:     NodeServer,
			MsgId:   32767,
			Len:     4294967295,
		},
		Body: data,
	}
	fmt.Println(msg.Head.Magic[0])
	//fmt.Println(recursionStruct(reflect.ValueOf(msg)))
	//t.Log(reflect.ValueOf(msg).Elem().Kind())
}

func recursionStruct(t reflect.Value) (data []byte) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		fmt.Println("no struct", t)
		return
	}
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Kind() == reflect.Struct || (t.Field(i).Kind() ==
			reflect.Ptr && t.Field(i).Elem().Kind() == reflect.Struct) {
			recursionStruct(t.Field(i))
			continue
		}
		fmt.Println(t.Field(i).Type())
	}
	return
}

func TestMaxInt16(t *testing.T) {
	assert.Equal(t, 32767, 1<<16/2-1)
}

func TestMsg_UnmarshalBinary(t *testing.T) {
	data := []byte{'X', 'B', 0, 15, 0, 2, 0, 1, 0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o'}
	msg := &Msg{}
	msg.UnmarshalBinary(data)
	t.Log(msg)
	t.Log(data)

	msg = &Msg{}
	b := bytes.NewBuffer(data)
	n, err := fmt.Fscan(b, &msg.Head.Magic, &msg.Head.Command, &msg.Head.Tag, &msg.Head.MsgId, &msg.Head.Len, &msg.Body)
	t.Log(n, err, msg)
}
