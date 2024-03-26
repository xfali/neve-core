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

	// 获得Bean生命周期相关的方法名
	BeanLifeCycleMethodNames() map[LifeCycle]string

	//Deprecated: BeanFactory返回值包含的初始化方法名，可为空
	InitMethodName() string

	//Deprecated: BeanFactory返回值包含的销毁方法名，可为空
	DestroyMethodName() string
}

type beanFactoryOpt func(*defaultCustomBeanFactory)

type defaultCustomBeanFactory struct {
	beanFunc       interface{}
	names          []string
	lifeCycleFuncs map[LifeCycle]string
}

func NewCustomBeanFactory(beanFunc interface{}, initMethod, destroyMethod string) *defaultCustomBeanFactory {
	return NewCustomBeanFactoryWithName(beanFunc, nil, initMethod, destroyMethod)
}

func NewCustomBeanFactoryWithName(beanFunc interface{}, names []string, initMethod, destroyMethod string) *defaultCustomBeanFactory {
	return NewCustomBeanFactoryWithOpts(beanFunc,
		CustomBeanFactoryOpts.Names(names),
		CustomBeanFactoryOpts.PostAfterSet(initMethod),
		CustomBeanFactoryOpts.PreDestroy(destroyMethod))
}

func NewCustomBeanFactoryWithOpts(beanFunc interface{}, opts ...beanFactoryOpt) *defaultCustomBeanFactory {
	ft := reflect.TypeOf(beanFunc)
	if err := verifyBeanFunctionEx(ft); err != nil {
		panic(fmt.Errorf("NewCustomMethodBean with a invalid function type: %s", ft.String()))
	}
	ret := &defaultCustomBeanFactory{
		beanFunc:       beanFunc,
		lifeCycleFuncs: map[LifeCycle]string{},
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func (b *defaultCustomBeanFactory) BeanFactory() interface{} {
	return b.beanFunc
}

func (b *defaultCustomBeanFactory) InjectNames() []string {
	return b.names
}

func (b *defaultCustomBeanFactory) BeanLifeCycleMethodNames() map[LifeCycle]string {
	return b.lifeCycleFuncs
}

func (b *defaultCustomBeanFactory) InitMethodName() string {
	return b.lifeCycleFuncs[PostAfterSet]
}

func (b *defaultCustomBeanFactory) DestroyMethodName() string {
	return b.lifeCycleFuncs[PreDestroy]
}

type customBeanFactoryOpts struct {
}

func (o customBeanFactoryOpts) Names(names []string) beanFactoryOpt {
	return func(factory *defaultCustomBeanFactory) {
		factory.names = names
	}
}

func (o customBeanFactoryOpts) AddLifeCycleMethod(cycle LifeCycle, methodName string) beanFactoryOpt {
	return func(factory *defaultCustomBeanFactory) {
		if methodName != "" {
			factory.lifeCycleFuncs[cycle] = methodName
		}
	}
}

func (o customBeanFactoryOpts) PostAfterSet(methodName string) beanFactoryOpt {
	return func(factory *defaultCustomBeanFactory) {
		if methodName != "" {
			factory.lifeCycleFuncs[PostAfterSet] = methodName
		}
	}
}

func (o customBeanFactoryOpts) PreAfterSet(methodName string) beanFactoryOpt {
	return func(factory *defaultCustomBeanFactory) {
		if methodName != "" {
			factory.lifeCycleFuncs[PreAfterSet] = methodName
		}
	}
}

func (o customBeanFactoryOpts) PostDestroy(methodName string) beanFactoryOpt {
	return func(factory *defaultCustomBeanFactory) {
		if methodName != "" {
			factory.lifeCycleFuncs[PostDestroy] = methodName
		}
	}
}

func (o customBeanFactoryOpts) PreDestroy(methodName string) beanFactoryOpt {
	return func(factory *defaultCustomBeanFactory) {
		if methodName != "" {
			factory.lifeCycleFuncs[PreDestroy] = methodName
		}
	}
}

var CustomBeanFactoryOpts customBeanFactoryOpts

type customMethodBeanDefinition struct {
	functionExDefinition

	lifeCycleFuncs map[LifeCycle]string
}

func newCustomMethodBeanDefinition(b CustomBeanFactory) (Definition, error) {
	d, err := newFunctionExDefinition(b.BeanFactory())
	if err != nil {
		return nil, err
	}
	ret := &customMethodBeanDefinition{
		functionExDefinition: *d.(*functionExDefinition),
		lifeCycleFuncs:       b.BeanLifeCycleMethodNames(),
	}

	return ret, ret.verifyCustomBeanFunction()
}

func checkPublic(name string) bool {
	return name[0] >= 'A' && name[0] <= 'Z'
}

func (d *customMethodBeanDefinition) verifyCustomBeanFunction() error {
	rt := d.t
	for k, v := range d.lifeCycleFuncs {
		if v != "" {
			if !checkPublic(v) {
				return fmt.Errorf("Type %s [%s] method %s is private ", reflection.GetTypeName(d.t), k, v)
			}
			m, ok := rt.MethodByName(v)
			if !ok {
				return fmt.Errorf("Type %s [%s] method %s not found ", reflection.GetTypeName(d.t), k, v)
			} else {
				if m.Type.NumIn() != 1 {
					return fmt.Errorf("Type %s [%s] method %s cannot with params ", reflection.GetTypeName(d.t), k, v)
				}
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
	return fmt.Errorf("%s method %s is invalid", reflection.GetTypeName(value.Type()), name)
}

func (d *customMethodBeanDefinition) AfterSet() error {
	if atomic.CompareAndSwapInt32(&d.initOnce, 0, 1) {
		d.instanceLock.RLock()
		defer d.instanceLock.RUnlock()
		var errs errors.Errors
		for _, i := range d.instances.MapKeys() {
			if i.IsValid() && !i.IsNil() {
				name := d.lifeCycleFuncs[PreAfterSet]
				if name != "" {
					err := d.callByName(i, name)
					if err != nil {
						_ = errs.AddError(err)
					}
				}
				if d.t.Implements(InitializingType) {
					err := i.Interface().(Initializing).BeanAfterSet()
					if err != nil {
						_ = errs.AddError(err)
					}
				}
				name = d.lifeCycleFuncs[PostAfterSet]
				if name != "" {
					err := d.callByName(i, name)
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
				name := d.lifeCycleFuncs[PreDestroy]
				if name != "" {
					err := d.callByName(i, name)
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

				name = d.lifeCycleFuncs[PostAfterSet]
				if name != "" {
					err := d.callByName(i, name)
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
