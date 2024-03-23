/*
 * Copyright 2022 Xiongfa Li.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package injector

import (
	"errors"
	"fmt"
	"github.com/xfali/neve-core/bean"
	"github.com/xfali/neve-core/reflection"
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

func NewDefaultInjectFunctionHandler(logger xlog.Logger, manager ListenerManager) *defaultInjectFunctionHandler {
	ret := &defaultInjectFunctionHandler{
		logger:  logger,
		creator: create,
	}
	ret.lm = manager
	return ret
}

func (fi *defaultInjectFunctionHandler) SetListenerManager(manager ListenerManager) {
	fi.lm = manager
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

// FIXME: This isn't a elegant implementation (but works well).
func WrapBean(o interface{}, container bean.Container, injector Injector, manager ListenerManager) (interface{}, error) {
	// 如果是CustomBeanFactory则需要将创建bean方法的参数自动代理注入，变为无参数仅返回的创建方法
	if b, ok := o.(bean.CustomBeanFactory); ok {
		fac := b.BeanFactory()
		if reflect.TypeOf(fac).NumIn() > 0 {
			names := b.InjectNames()
			if len(names) > 0 {
				f, err := WrapBeanFactoryByNameFunc(fac, names, container, injector, manager)
				if err != nil {
					return nil, err
				}
				return bean.NewCustomBeanFactory(f, b.InitMethodName(), b.DestroyMethodName()), nil
			} else {
				f, err := WrapBeanFactoryFunc(fac, container, injector, manager)
				if err != nil {
					return nil, err
				}
				return bean.NewCustomBeanFactory(f, b.InitMethodName(), b.DestroyMethodName()), nil
			}
		} else {
			return o, nil
		}
	}
	return WrapBeanFactoryFunc(o, container, injector, manager)
}
