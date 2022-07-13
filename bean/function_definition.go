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
	"errors"
	"fmt"
	errors2 "github.com/xfali/neve-core/errors"
	"github.com/xfali/neve-core/reflection"
	"reflect"
	"sync"
	"sync/atomic"
)

const (
	functionDefinitionNone = iota
	functionDefinitionInjecting
)

var (
	DummyType  = reflect.TypeOf((*struct{})(nil)).Elem()
	DummyValue = reflect.ValueOf(struct{}{})
)

type functionExDefinition struct {
	name   string
	o      interface{}
	fn     reflect.Value
	t      reflect.Type
	status int32

	instances    reflect.Value
	instanceLock sync.RWMutex
	initOnce     int32
	destroyOnce  int32
}

func verifyBeanFunctionEx(ft reflect.Type) error {
	if ft.Kind() != reflect.Func {
		return errors.New("Param not function ")
	}
	if ft.NumOut() != 1 {
		return errors.New("Bean function must ONLY have 1 return value ")
	}

	rt := ft.Out(0)
	if rt.Kind() != reflect.Ptr && rt.Kind() != reflect.Interface {
		return errors.New("Bean function 1st return value must be pointer or interface ")
	}

	return nil
}

func newFunctionExDefinition(o interface{}) (Definition, error) {
	ft := reflect.TypeOf(o)
	err := verifyBeanFunctionEx(ft)
	if err != nil {
		return nil, err
	}
	ot := ft.Out(0)
	fn := reflect.ValueOf(o)
	ret := &functionExDefinition{
		o:         o,
		name:      reflection.GetTypeName(ot),
		fn:        fn,
		t:         ot,
		instances: reflect.MakeMap(reflect.MapOf(ot, DummyType)),
	}
	return ret, nil
}

func (d *functionExDefinition) Type() reflect.Type {
	return d.t
}

func (d *functionExDefinition) Name() string {
	return d.name
}

func (d *functionExDefinition) Value() reflect.Value {
	if atomic.CompareAndSwapInt32(&d.status, functionDefinitionNone, functionDefinitionInjecting) {
		defer atomic.CompareAndSwapInt32(&d.status, functionDefinitionInjecting, functionDefinitionNone)
		v := d.fn.Call(nil)[0]
		if v.IsValid() {
			d.instanceLock.Lock()
			defer d.instanceLock.Unlock()
			d.instances.SetMapIndex(v, DummyValue)
		}
		return v
	} else {
		panic(fmt.Errorf("BeanDefinition: [Function] inject type [%s] Circular dependency ", d.name))
	}
}

func (d *functionExDefinition) Interface() interface{} {
	return d.o
}

func (d *functionExDefinition) IsObject() bool {
	return false
}

func (d *functionExDefinition) AfterSet() error {
	if atomic.CompareAndSwapInt32(&d.initOnce, 0, 1) {
		d.instanceLock.RLock()
		defer d.instanceLock.RUnlock()
		var errs errors2.Errors

		for _, i := range d.instances.MapKeys() {
			if i.IsValid() && !i.IsNil() {
				if v, ok := i.Interface().(Initializing); ok {
					err := v.BeanAfterSet()
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

func (d *functionExDefinition) Destroy() error {
	if atomic.CompareAndSwapInt32(&d.destroyOnce, 0, 1) {
		d.instanceLock.RLock()
		defer d.instanceLock.RUnlock()
		var errs errors2.Errors
		for _, i := range d.instances.MapKeys() {
			if i.IsValid() && !i.IsNil() {
				if v, ok := i.Interface().(Disposable); ok {
					err := v.BeanDestroy()
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

func (d *functionExDefinition) Classify(classifier Classifier) (bool, error) {
	d.instanceLock.RLock()
	defer d.instanceLock.RUnlock()
	var errs errors2.Errors
	ok := false
	for _, i := range d.instances.MapKeys() {
		if !i.IsNil() {
			ret, err := classifier.Classify(i.Interface())
			if ret {
				ok = ret
			}
			if err != nil {
				_ = errs.AddError(err)
			}
		}
	}
	if errs.Empty() {
		return ok, nil
	}
	return ok, errs
}

type functionDefinition struct {
	name string
	o    interface{}
	fn   reflect.Value
	t    reflect.Type
}

func verifyBeanFunction(ft reflect.Type) error {
	if ft.Kind() != reflect.Func {
		return errors.New("Param not function. ")
	}
	if ft.NumOut() != 1 {
		return errors.New("Bean function must have 1 return value: TYPE. ")
	}

	rt := ft.Out(0)
	if rt.Kind() != reflect.Ptr && rt.Kind() != reflect.Interface {
		return errors.New("Bean function 1st return value must be pointer or interface. ")
	}
	return nil
}

//Deprecated
func newFunctionDefinition(o interface{}) (Definition, error) {
	ft := reflect.TypeOf(o)
	err := verifyBeanFunction(ft)
	if err != nil {
		return nil, err
	}
	ot := ft.Out(0)
	fn := reflect.ValueOf(o)
	return &functionDefinition{
		o:    o,
		name: reflection.GetTypeName(ot),
		fn:   fn,
		t:    ot,
	}, nil
}

func (d *functionDefinition) Type() reflect.Type {
	return d.t
}

func (d *functionDefinition) Name() string {
	return d.name
}

func (d *functionDefinition) Value() reflect.Value {
	return d.fn.Call(nil)[0]
}

func (d *functionDefinition) Interface() interface{} {
	return d.o
}

func (d *functionDefinition) IsObject() bool {
	return false
}

func (d *functionDefinition) AfterSet() error {
	return nil
}

func (d *functionDefinition) Destroy() error {
	return nil
}

func (d *functionDefinition) Classify(classifier Classifier) (bool, error) {
	return false, nil
}
