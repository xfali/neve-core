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

package bean

import (
	"fmt"
	"github.com/xfali/neve-core/errors"
	"github.com/xfali/neve-core/reflection"
	"reflect"
	"sync"
	"sync/atomic"
)

type CustomBeanFactory interface {
	// 返回或者创建bean的方法
	// 该方法可能包含一个或者多个参数，参数会在实例化时自动注入
	// 该方法只能有一个返回值，返回的值将被注入到依赖该类型值的对象中
	BeanFactory() interface{}

	// BeanFactory返回创建bean方法如果带参数，且参数需要指定注入名称时将根据InjectNames返回的名称列表进行匹配
	// 注意：
	// 1、如果所有参数都不需要名称匹配，则返回nil
	// 2、如果需要使用名称匹配则：返回的string数组长度需要与创建bean方法的常数个数一致
	// 3、如果需要部分匹配，则需要自动匹配的参数对应的name填入空字符串""
	InjectNames() []string

	// BeanFactory返回值包含的初始化方法名，可为空
	InitMethodName() string

	// BeanFactory返回值包含的销毁方法名，可为空
	DestroyMethodName() string
}

type defaultCustomBeanFactory struct {
	beanFunc      interface{}
	names         []string
	initMethod    string
	destroyMethod string
}

func NewCustomBeanFactory(beanFunc interface{}, initMethod, destroyMethod string) *defaultCustomBeanFactory {
	ft := reflect.TypeOf(beanFunc)
	if err := verifyBeanFunctionEx(ft); err != nil {
		panic(fmt.Errorf("NewCustomMethodBean with a invalid function type: %s, error: %v", ft.String(), err))
	}
	return &defaultCustomBeanFactory{
		beanFunc:      beanFunc,
		initMethod:    initMethod,
		destroyMethod: destroyMethod,
	}
}

func NewCustomBeanFactoryWithName(beanFunc interface{}, names []string, initMethod, destroyMethod string) *defaultCustomBeanFactory {
	ft := reflect.TypeOf(beanFunc)
	if err := verifyBeanFunctionEx(ft); err != nil {
		panic(fmt.Errorf("NewCustomMethodBean with a invalid function type: %s", ft.String()))
	}
	return &defaultCustomBeanFactory{
		beanFunc:      beanFunc,
		names:         names,
		initMethod:    initMethod,
		destroyMethod: destroyMethod,
	}
}

func (b *defaultCustomBeanFactory) BeanFactory() interface{} {
	return b.beanFunc
}

func (b *defaultCustomBeanFactory) InjectNames() []string {
	return b.names
}

func (b *defaultCustomBeanFactory) InitMethodName() string {
	return b.initMethod
}

func (b *defaultCustomBeanFactory) DestroyMethodName() string {
	return b.destroyMethod
}

type customMethodBeanDefinition struct {
	functionExDefinition
	initializingFuncName string
	disposableFuncName   string
}

func newCustomMethodBeanDefinition(b CustomBeanFactory) (Definition, error) {
	d, err := newFunctionExDefinition(b.BeanFactory())
	if err != nil {
		return nil, err
	}
	ret := &customMethodBeanDefinition{
		functionExDefinition: *d.(*functionExDefinition),
		initializingFuncName: b.InitMethodName(),
		disposableFuncName:   b.DestroyMethodName(),
	}

	return ret, ret.verifyCustomBeanFunction()
}

func checkPublic(name string) bool {
	return name[0] >= 'A' && name[0] <= 'Z'
}

func (d *customMethodBeanDefinition) verifyCustomBeanFunction() error {
	rt := d.t
	if d.initializingFuncName != "" {
		if !checkPublic(d.initializingFuncName) {
			return fmt.Errorf("Type %s init method %s is private ", reflection.GetTypeName(d.t), d.initializingFuncName)
		}
		m, ok := rt.MethodByName(d.initializingFuncName)
		if !ok {
			return fmt.Errorf("Type %s init method %s not found ", reflection.GetTypeName(d.t), d.initializingFuncName)
		} else {
			if m.Type.NumIn() == 0 {
				return fmt.Errorf("Type %s init method %s cannot with params ", reflection.GetTypeName(d.t), d.initializingFuncName)
			}
		}
	}

	if d.disposableFuncName != "" {
		if !checkPublic(d.initializingFuncName) {
			return fmt.Errorf("Type %s destroy method %s is private ", reflection.GetTypeName(d.t), d.initializingFuncName)
		}
		m, ok := rt.MethodByName(d.disposableFuncName)
		if !ok {
			return fmt.Errorf("Type %s destroy method %s not found ", reflection.GetTypeName(d.t), d.disposableFuncName)
		} else {
			if m.Type.NumIn() == 0 {
				return fmt.Errorf("Type %s destroy method %s cannot with params ", reflection.GetTypeName(d.t), d.disposableFuncName)
			}
		}
	}

	return nil
}

func (d *customMethodBeanDefinition) callByName(value reflect.Value, name string) error {
	m := value.MethodByName(name)
	if m.IsValid() && !m.IsNil() {
		// validate function must before newCustomMethodBeanDefinition!
		rets := m.Call(nil)
		for i := len(rets) - 1; i >= 0; i-- {
			ret := rets[i]
			if ret.IsValid() && ret.Type().Implements(ErrorType) {
				if !ret.IsNil() {
					return rets[i].Interface().(error)
				}
			}
		}
		return nil
	}
	return fmt.Errorf("%s method %s is invalid", reflection.GetTypeName(value.Type()), d.initializingFuncName)
}

func (d *customMethodBeanDefinition) AfterSet() error {
	if atomic.CompareAndSwapInt32(&d.initOnce, 0, 1) {
		d.instanceLock.RLock()
		defer d.instanceLock.RUnlock()
		var errs errors.Errors
		for _, i := range d.instances.MapKeys() {
			if i.IsValid() && !i.IsNil() {
				if d.t.Implements(InitializingType) {
					err := i.Interface().(Initializing).BeanAfterSet()
					if err != nil {
						_ = errs.AddError(err)
					}
				}
				if d.initializingFuncName != "" {
					err := d.callByName(i, d.initializingFuncName)
					if err != nil {
						_ = errs.AddError(err)
					}
				}
			}
		}
		if errs.Empty() {
			return nil
		}
		return errs
	}
	return nil
}

func (d *customMethodBeanDefinition) Destroy() error {
	if atomic.CompareAndSwapInt32(&d.destroyOnce, 0, 1) {
		d.instanceLock.RLock()
		defer d.instanceLock.RUnlock()
		var errs errors.Errors
		for _, i := range d.instances.MapKeys() {
			if i.IsValid() && !i.IsNil() {
				if d.disposableFuncName != "" {
					err := d.callByName(i, d.disposableFuncName)
					if err != nil {
						_ = errs.AddError(err)
					}
				}
				if d.t.Implements(DisposableType) {
					err := i.Interface().(Disposable).BeanDestroy()
					if err != nil {
						_ = errs.AddError(err)
					}
				}
			}
		}
		if errs.Empty() {
			return nil
		}
		return errs
	}
	return nil
}

type singletonFunction struct {
	once sync.Once
	f    interface{}

	ret []reflect.Value
}

func (f *singletonFunction) get() interface{} {
	ft := reflect.TypeOf(f.f)
	if ft.Kind() != reflect.Func {
		panic("Input Type is not a function. ")
	}
	if ft.NumOut() != 1 {
		panic("Input function must ONLY have 1 return value. ")
	}

	rt := ft.Out(0)
	if rt.Kind() != reflect.Ptr && rt.Kind() != reflect.Interface {
		panic("Bean function 1st return value must be pointer or interface. ")
	}

	retFv := reflect.MakeFunc(ft, func(args []reflect.Value) (results []reflect.Value) {
		f.once.Do(func() {
			fv := reflect.ValueOf(f.f)
			f.ret = fv.Call(args)
		})
		return f.ret
	})
	return retFv.Interface()
}

func SingletonFactory(function interface{}) interface{} {
	s := singletonFunction{
		f: function,
	}
	return s.get()
}
