package utils

import (
	"errors"
	"reflect"
)

type Function struct {
	Name string
	Func reflect.Value
	Args []interface{}
}

func (f *Function) Call(args ...interface{}) (interface{}, error) {
	if len(f.Args) > len(args) {
		return nil, errors.New("args not enough")
	}
	argsValue := make([]reflect.Value, len(f.Args))
	for i := 0; i < len(f.Args); i++ {
		argsValue[i] = reflect.ValueOf(args[i])
	}
	call := f.Func.Call(argsValue)
	return call[0].Interface(), nil
}

var FuncMap = make(map[string]*Function)

func InitFunction(obj interface{}) {
	valueOf := reflect.ValueOf(obj)
	typeOf := reflect.TypeOf(obj)
	for i := 0; i < typeOf.NumMethod(); i++ {
		methodName := typeOf.Method(i).Name
		//TODO some method do not add
		methodType := typeOf.Method(i).Type
		f := &Function{
			Name: methodName,
			Func: valueOf.MethodByName(methodName),
			Args: make([]interface{}, methodType.NumIn()-1),
		}
		for j := 0; j > methodType.NumIn()-1; j++ {
			f.Args[i] = methodType.In(j).Kind()
		}
		FuncMap[methodName] = f
	}
}
