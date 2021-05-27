// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package appcontext

import (
	"errors"
	"fmt"
	"github.com/xfali/neve-core/bean"
	"github.com/xfali/neve-core/injector"
	"github.com/xfali/neve-utils/reflection"
	"github.com/xfali/xlog"
	"reflect"
)

type defaultInjectInvoker struct {
	types []reflect.Type
	names []string
	fv    reflect.Value
}

func (invoker *defaultInjectInvoker) Invoke(injector injector.Injector, container bean.Container) error {
	values := make([]reflect.Value, len(invoker.types))
	haveName := len(invoker.names) > 0
	for i, t := range invoker.types {
		o := reflect.New(t).Elem()
		name := ""
		if haveName {
			name = invoker.names[i]
		}
		err := injector.InjectValue(container, name, o)
		if err != nil {
			return err
		}
		values[i] = o
	}

	invoker.fv.Call(values)
	return nil
}

func (invoker *defaultInjectInvoker) ResolveFunction(injector injector.Injector, names []string, function interface{}) error {
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
			return errors.New("Names not match function's params. ")
		} else {
			invoker.names = names
		}
	}

	for i := 0; i < s; i++ {
		t := t.In(i)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if !injector.CanInjectType(t) {
			return fmt.Errorf("Cannot Inject Type : %s . ", reflection.GetTypeName(t))
		}
		invoker.types = append(invoker.types, t)
	}
	invoker.fv = reflect.ValueOf(function)
	return nil
}

type defaultInjectFunctionHandler struct {
	logger   xlog.Logger
	injector injector.Injector
	creator  func() FunctionInjectInvoker

	invokers []FunctionInjectInvoker
}

func newDefaultInjectFunctionHandler(logger xlog.Logger) *defaultInjectFunctionHandler {
	ret := &defaultInjectFunctionHandler{
		logger:  logger,
		creator: create,
	}
	return ret
}

func (fi *defaultInjectFunctionHandler) SetInjector(injector injector.Injector) {
	fi.injector = injector
}

func (fi *defaultInjectFunctionHandler) InjectAllFunctions(container bean.Container) error {
	var last error
	for _, invoker := range fi.invokers {
		err := invoker.Invoke(fi.injector, container)
		if err != nil {
			fi.logger.Errorf("Inject function failed: %s error: %s\n", err)
			last = err
		}
	}
	return last
}

func create() FunctionInjectInvoker {
	return &defaultInjectInvoker{}
}

func (fi *defaultInjectFunctionHandler) RegisterInjectFunction(function interface{}) error {
	return fi.RegisterInjectFunctionWithNames(nil, function)
}

func (fi *defaultInjectFunctionHandler) RegisterInjectFunctionWithNames(names []string, function interface{}) error {
	invoker := fi.creator()
	if err := invoker.ResolveFunction(fi.injector, names, function); err != nil {
		return err
	}
	fi.invokers = append(fi.invokers, invoker)
	return nil
}
