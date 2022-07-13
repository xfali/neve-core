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
	"sync/atomic"
)

type CustomMethodBean interface {
	// 返回或者创建bean的方法
	// 该方法可能包含一个或者多个参数，参数会在实例化时自动注入
	// 该方法只能有一个返回值，返回的值将被注入到依赖该类型值的对象中
	BeanFactory() interface{}

	// BeanFactory返回值包含的初始化方法名，可为空
	InitMethodName() string

	// BeanFactory返回值包含的销毁方法名，可为空
	DestroyMethodName() string
}

type defaultCustomMethodBean struct {
	beanFunc      interface{}
	initMethod    string
	destroyMethod string
}

func NewCustomMethodBean(beanFunc interface{}, initMethod, destroyMethod string) *defaultCustomMethodBean {
	ft := reflect.TypeOf(beanFunc)
	if err := verifyBeanFunctionEx(ft); err != nil {
		panic(fmt.Errorf("NewCustomMethodBean with a invalid function type: %s", ft.String()))
	}
	return &defaultCustomMethodBean{
		beanFunc:      beanFunc,
		initMethod:    initMethod,
		destroyMethod: destroyMethod,
	}
}

func (b *defaultCustomMethodBean) BeanFactory() interface{} {
	return b.beanFunc
}

func (b *defaultCustomMethodBean) InitMethodName() string {
	return b.initMethod
}

func (b *defaultCustomMethodBean) DestroyMethodName() string {
	return b.destroyMethod
}

type customMethodBeanDefinition struct {
	functionExDefinition
	initializingFuncName string
	disposableFuncName   string
}

func newCustomMethodBeanDefinition(b CustomMethodBean) (Definition, error) {
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
		for _, i := range d.instances {
			if i.IsValid() && !i.IsNil() {
				if d.initializingFuncName == "" {
					if d.t.Implements(InitializingType) {
						err := i.Interface().(Initializing).BeanAfterSet()
						if err != nil {
							_ = errs.AddError(err)
						}
					}
				} else {
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
		for _, i := range d.instances {
			if i.IsValid() && !i.IsNil() {
				if d.disposableFuncName == "" {
					if d.t.Implements(DisposableType) {
						err := i.Interface().(Disposable).BeanDestroy()
						if err != nil {
							_ = errs.AddError(err)
						}
					}
				} else {
					err := d.callByName(i, d.disposableFuncName)
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
