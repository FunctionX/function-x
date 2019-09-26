package p2p

import (
	"context"
	"reflect"
	"sync"

	"fx/chain/logger"
)

const subscribeLatest = "subscribeLatest"
const subscribeFrom = "subscribeFrom"

type subFunc struct {
	funcName string
	name     string
	argsJson string
	callback EventCallback
	sub      *Subscriber
}
type Subscriber struct {
	*Result
	sub      interface{}
	ch       chan struct{}
	callback EventCallback
	decode   *Contract
	name     string
	close    func()
	ctx      context.Context
	cancel   context.CancelFunc
}

func (sub *Subscriber) Stop() {
	handler.cancelHandler(sub)
	if sub.close != nil {
		sub.close()
	}
	sub.cancel()
	close(sub.ch)
}

func (sub *Subscriber) closeSub() {
	if sub.close != nil {
		sub.close()
	}
	sub.cancel()
	close(sub.ch)
}

//name(key)-type(value)
type args struct {
	names    []string //ordered
	formates map[string]argType
}

// Type is the reflection of the supported argument type
type argType struct {
	Elem *argType

	Kind reflect.Kind
	Type reflect.Type
	Size int
	T    byte // Our own type checking

	stringKind string // holds the unparsed string for deriving signatures
}

//method defined
type method struct {
	name   string
	input  args
	output args
}

type Contract struct {
	id        string
	calls     map[string]method //read-only methods
	transacts map[string]method //write-only methods
	events    map[string]method //event methods
}

var handler = &contractHandler{handlers: make(map[*Contract][]subFunc)}

type contractHandler struct {
	handlers map[*Contract][]subFunc
	sync.Mutex
}

func (c *contractHandler) registerHandler(contract *Contract, subFunc1 subFunc) {
	c.Lock()
	defer c.Unlock()
	for i := range c.handlers[contract] {
		if c.handlers[contract][i].name == subFunc1.name {
			logger.Warn("sub func exist", "id", contract.id, "name", subFunc1.name)
			c.handlers[contract][i].sub.closeSub()
			c.handlers[contract][i] = subFunc1
			return
		}
	}
	c.handlers[contract] = append(c.handlers[contract], subFunc1)
}

func (c *contractHandler) getSubFunc(contract *Contract, name string) subFunc {
	c.Lock()
	defer c.Unlock()
	for i := range c.handlers[contract] {
		if c.handlers[contract][i].name == name {
			return c.handlers[contract][i]
		}
	}
	return subFunc{}
}

func (c *contractHandler) cancelHandler(sub *Subscriber) {
	c.Lock()
	defer c.Unlock()
	subFuncs, ok := c.handlers[sub.decode]
	if !ok {
		logger.Error("contract handler not exist sub", "info", sub.name)
		return
	}
	for i := 0; i < len(subFuncs); i++ {
		if sub.name == subFuncs[i].name && sub.callback == subFuncs[i].callback {
			subFuncs = append(subFuncs[:i], subFuncs[i+1:]...)
		}
	}
	c.handlers[sub.decode] = subFuncs
	logger.Debug("contract cancel handler", "sub", sub)
	if len(subFuncs) <= 0 {
		delete(c.handlers, sub.decode)
	}
}

func (c *contractHandler) doHandler() {
	c.Lock()
	defer c.Unlock()
	for contract, subFuncs := range c.handlers {
		logger.Debug("doHandler", "contract", contract.id, "subFuncs", len(subFuncs))
		for i := 0; i < len(subFuncs); i++ {
			if subFuncs[i].sub == nil {
				continue // none sub
			}
			sf := subFuncs[i] // cache
			sf.sub.closeSub()
			subFuncs = append(subFuncs[:i], subFuncs[i+1:]...)
			var err error
			switch sf.funcName {
			case subscribeLatest, subscribeFrom:
				logger.Debug("doHandler subscribeLatest")
				go func(c *Contract, f subFunc) {
					if _, err = c.subscribeLatest(f.name, f.argsJson, f.callback); err != nil {
						logger.Error("contract handler subscribeLatest", "err", err.Error())
					} else {
						logger.Info("success start", "subFuncName", f.name)
					}
				}(contract, sf)
			default:
				logger.Error("subscribe not exist", "func name", sf.funcName)
				continue
			}
		}
		c.handlers[contract] = subFuncs
	}
}

func (c *Contract) subscribeLatest(name string, argsJson string, callback EventCallback) (*Subscriber, error) {
	return nil, nil
}
