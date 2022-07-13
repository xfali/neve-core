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

type Classifier interface {
	// 对象分类，判断对象是否实现某些接口，并进行相关归类。为了支持多协程处理，该方法应线程安全。
	// 注意：该方法建议只做归类，具体处理使用Process，不保证Processor的实现在此方法中做了相关处理。
	// 该方法在Bean Inject注入之后调用
	// return: bool 是否能够处理对象， error 处理是否有错误
	Classify(o interface{}) (bool, error)
}

type Definition interface {
	// 类型
	Type() reflect.Type

	// 名称
	Name() string

	// 获得值
	Value() reflect.Value

	// 获得注册的bean对象
	Interface() interface{}

	// 是否是可注入对象
	IsObject() bool

	// 在属性配置完成后调用
	AfterSet() error

	// 销毁对象
	Destroy() error

	// 对对象进行分类
	Classify(classifier Classifier) (bool, error)
}

type objectDefinition struct {
	name        string
	o           interface{}
	t           reflect.Type
	flagSet     int32
	flagDestroy int32
}

type DefinitionCreator func(o interface{}) (Definition, error)

var (
	InitializingType       = reflect.TypeOf((*Initializing)(nil)).Elem()
	DisposableType         = reflect.TypeOf((*Disposable)(nil)).Elem()
	ErrorType              = reflect.TypeOf((*error)(nil)).Elem()
	beanDefinitionCreators = map[reflect.Kind]DefinitionCreator{
		reflect.Ptr:   newObjectDefinition,
		reflect.Func:  newFunctionExDefinition,
		reflect.Slice: newSliceDefinition,
		reflect.Map:   newMapDefinition,
	}
)

// 注册BeanDefinition创建器，使其能处理更多类型。
// 默认支持Pointer、Function
func RegisterBeanDefinitionCreator(kind reflect.Kind, creator DefinitionCreator) {
	if creator != nil {
		beanDefinitionCreators[kind] = creator
	}
}

func CreateBeanDefinition(o interface{}) (Definition, error) {
	if v, ok := o.(CustomMethodBean); ok {
		return newCustomMethodBeanDefinition(v)
	}

	t := reflect.TypeOf(o)
	creator, ok := beanDefinitionCreators[t.Kind()]
	if !ok || creator == nil {
		return nil, fmt.Errorf("Cannot handle this type: %s" + t.String())
	}

	return creator(o)
}

func newObjectDefinition(o interface{}) (Definition, error) {
	t := reflect.TypeOf(o)
	if t.Kind() == reflect.Ptr {
		t2 := t.Elem()
		if t2.Kind() == reflect.Ptr {
			return nil, errors.New("o must be a Pointer but get Pointer's Pointer")
		}
	}
	return &objectDefinition{
		name:        reflection.GetTypeName(t),
		o:           o,
		t:           t,
		flagSet:     0,
		flagDestroy: 0,
	}, nil
}

func (d *objectDefinition) Type() reflect.Type {
	return d.t
}

func (d *objectDefinition) Name() string {
	return d.name
}

func (d *objectDefinition) Value() reflect.Value {
	return reflect.ValueOf(d.o)
}

func (d *objectDefinition) Interface() interface{} {
	return d.o
}

func (d *objectDefinition) IsObject() bool {
	return true
}

func (d *objectDefinition) AfterSet() error {
	// Just run once
	if atomic.CompareAndSwapInt32(&d.flagSet, 0, 1) {
		if v, ok := d.o.(Initializing); ok {
			return v.BeanAfterSet()
		}
	}
	return nil
}

func (d *objectDefinition) Destroy() error {
	// Just run once
	if atomic.CompareAndSwapInt32(&d.flagDestroy, 0, 1) {
		if v, ok := d.o.(Disposable); ok {
			return v.BeanDestroy()
		}
	}
	return nil
}

func (d *objectDefinition) Classify(classifier Classifier) (bool, error) {
	return classifier.Classify(d.o)
}

var errType = reflect.TypeOf((*error)(nil)).Elem()

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

const (
	functionDefinitionNone = iota
	functionDefinitionInjecting
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

func NewCustomMethodBean(beanFunc interface{}, initMethod, destoryMethod string) *defaultCustomMethodBean {
	ft := reflect.TypeOf(beanFunc)
	if err := verifyBeanFunctionEx(ft); err != nil {
		panic(fmt.Errorf("NewCustomMethodBean with a invalid function type: %s", ft.String()))
	}
	return &defaultCustomMethodBean{
		beanFunc:      beanFunc,
		initMethod:    initMethod,
		destroyMethod: destoryMethod,
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
	if d.initializingFuncName == "" {
		return nil
	}
	if atomic.CompareAndSwapInt32(&d.initOnce, 0, 1) {
		d.instanceLock.RLock()
		defer d.instanceLock.RUnlock()
		var errs errors2.Errors
		for _, i := range d.instances {
			if i.IsValid() && !i.IsNil() {
				err := d.callByName(i, d.initializingFuncName)
				if err != nil {
					_ = errs.AddError(err)
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
	if d.disposableFuncName == "" {
		return nil
	}
	if atomic.CompareAndSwapInt32(&d.destroyOnce, 0, 1) {
		d.instanceLock.RLock()
		defer d.instanceLock.RUnlock()
		var errs errors2.Errors
		for _, i := range d.instances {
			if i.IsValid() && !i.IsNil() {
				err := d.callByName(i, d.disposableFuncName)
				if err != nil {
					_ = errs.AddError(err)
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

type functionExDefinition struct {
	name   string
	o      interface{}
	fn     reflect.Value
	t      reflect.Type
	status int32

	instances    []reflect.Value
	instanceLock sync.RWMutex
	initOnce     int32
	destroyOnce  int32
}

func verifyBeanFunctionEx(ft reflect.Type) error {
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

func newFunctionExDefinition(o interface{}) (Definition, error) {
	ft := reflect.TypeOf(o)
	err := verifyBeanFunctionEx(ft)
	if err != nil {
		return nil, err
	}
	ot := ft.Out(0)
	fn := reflect.ValueOf(o)
	ret := &functionExDefinition{
		o:    o,
		name: reflection.GetTypeName(ot),
		fn:   fn,
		t:    ot,
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
			d.instances = append(d.instances, v)
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
		for _, i := range d.instances {
			if !i.IsNil() {
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
		for _, i := range d.instances {
			if !i.IsNil() {
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
	for _, i := range d.instances {
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

type sliceDefinition struct {
	name string
	o    interface{}
	t    reflect.Type
}

func newSliceDefinition(o interface{}) (Definition, error) {
	t := reflect.TypeOf(o)
	return &sliceDefinition{
		name: reflection.GetSliceName(t),
		o:    o,
		t:    t,
	}, nil
}

func (d *sliceDefinition) Type() reflect.Type {
	return d.t
}

func (d *sliceDefinition) Name() string {
	return d.name
}

func (d *sliceDefinition) Value() reflect.Value {
	return reflect.ValueOf(d.o)
}

func (d *sliceDefinition) Interface() interface{} {
	return d.o
}

func (d *sliceDefinition) IsObject() bool {
	return false
}

func (d *sliceDefinition) AfterSet() error {
	return nil
}

func (d *sliceDefinition) Destroy() error {
	return nil
}

func (d *sliceDefinition) Classify(classifier Classifier) (bool, error) {
	return false, nil
}

type mapDefinition struct {
	name string
	o    interface{}
	t    reflect.Type
}

func newMapDefinition(o interface{}) (Definition, error) {
	t := reflect.TypeOf(o)
	return &mapDefinition{
		name: reflection.GetMapName(t),
		o:    o,
		t:    t,
	}, nil
}

func (d *mapDefinition) Type() reflect.Type {
	return d.t
}

func (d *mapDefinition) Name() string {
	return d.name
}

func (d *mapDefinition) Value() reflect.Value {
	return reflect.ValueOf(d.o)
}

func (d *mapDefinition) Interface() interface{} {
	return d.o
}

func (d *mapDefinition) IsObject() bool {
	return false
}

func (d *mapDefinition) AfterSet() error {
	return nil
}

func (d *mapDefinition) Destroy() error {
	return nil
}

func (d *mapDefinition) Classify(classifier Classifier) (bool, error) {
	return false, nil
}
