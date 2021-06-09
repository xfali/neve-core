// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package injector

import (
	"errors"
	"fmt"
	"github.com/xfali/neve-core/bean"
	"github.com/xfali/neve-utils/reflection"
	"github.com/xfali/xlog"
	"reflect"
	"sync"
)

type defaultInjectInvoker struct {
	types    []reflect.Type
	names    []string
	fv       reflect.Value
	funcName string
}

func (invoker *defaultInjectInvoker) Invoke(ij Injector, container bean.Container, manager ListenerManager) error {
	values := make([]reflect.Value, len(invoker.types))
	haveName := len(invoker.names) > 0
	for i, t := range invoker.types {
		o := reflect.New(t).Elem()
		name := ""
		var listeners []Listener
		if haveName {
			name = invoker.names[i]
		}
		if manager != nil {
			name, listeners = manager.ParseListener(name)
		}
		err := ij.InjectValue(container, name, o)
		if err != nil {
			err = fmt.Errorf("Inject function [%s] failed:error: %s\n", invoker.FunctionName(), err.Error())
			for _, l := range listeners {
				l.OnInjectFailed(err)
			}
			return err
		}
		values[i] = o
	}

	invoker.fv.Call(values)
	return nil
}

func (invoker *defaultInjectInvoker) FunctionName() string {
	return invoker.funcName
}

func (invoker *defaultInjectInvoker) ResolveFunction(injector Injector, names []string, function interface{}) error {
	t := reflect.TypeOf(function)
	if t.Kind() != reflect.Func {
		return errors.New("Param is not a function. ")
	}

	s := t.NumIn()
	if s == 0 {
		return errors.New("Param is not match, expect func(Type1, Type2...TypeN). ")
	}

	if len(names) > 0 {
		if len(names) != s {
			//return errors.New("Names not match function's params. ")
		}
		invoker.names = formatNames(names, s)
	}

	for i := 0; i < s; i++ {
		tt := t.In(i)
		if !injector.CanInjectType(tt) {
			return fmt.Errorf("Cannot Inject Type : %s . ", reflection.GetTypeName(tt))
		}
		invoker.types = append(invoker.types, tt)
	}
	invoker.fv = reflect.ValueOf(function)
	invoker.funcName = reflection.GetTypeName(invoker.fv.Type())
	if invoker.funcName == "" {
		invoker.funcName = "func"
	}
	return nil
}

func formatNames(names []string, size int) []string {
	srcSize := len(names)
	if srcSize == size {
		return names
	} else if srcSize > size {
		return names[:size]
	} else {
		for i := srcSize; i < size; i++ {
			names = append(names, "")
		}
		return names
	}
}

type defaultInjectFunctionHandler struct {
	logger   xlog.Logger
	injector Injector
	creator  func() FunctionInjectInvoker

	lm       ListenerManager
	invokers []FunctionInjectInvoker
	locker   sync.Mutex
}

func NewDefaultInjectFunctionHandler(logger xlog.Logger) *defaultInjectFunctionHandler {
	ret := &defaultInjectFunctionHandler{
		logger:  logger,
		creator: create,
	}
	ret.lm = NewListenerManager(ret.logger)
	return ret
}

func (fi *defaultInjectFunctionHandler) SetInjector(injector Injector) {
	fi.injector = injector
}

func (fi *defaultInjectFunctionHandler) InjectAllFunctions(container bean.Container) error {
	var last error

	fi.locker.Lock()
	defer fi.locker.Unlock()

	for _, invoker := range fi.invokers {
		err := invoker.Invoke(fi.injector, container, fi.lm)
		if err != nil {
			//fi.logger.Errorf("Inject function failed: %s error: %s\n", invoker.FunctionName(), err.Error())
			last = err
		}
	}
	return last
}

func create() FunctionInjectInvoker {
	return &defaultInjectInvoker{}
}

func (fi *defaultInjectFunctionHandler) RegisterInjectFunction(function interface{}, names ...string) error {
	invoker := fi.creator()
	if err := invoker.ResolveFunction(fi.injector, names[:], function); err != nil {
		return err
	}
	fi.addInvoker(invoker)
	return nil
}

func (fi *defaultInjectFunctionHandler) addInvoker(invoker FunctionInjectInvoker) {
	fi.locker.Lock()
	defer fi.locker.Unlock()

	fi.invokers = append(fi.invokers, invoker)
}