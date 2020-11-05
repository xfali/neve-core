// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package bean

import (
	"errors"
	"github.com/xfali/neve-utils/reflection"
	"reflect"
	"sync/atomic"
)

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
}

type objectDefinition struct {
	name        string
	o           interface{}
	t           reflect.Type
	flagSet     int32
	flagDestroy int32
}

type DefinitionCreator func(o interface{}) (Definition, error)

var beanDefinitionCreators = map[reflect.Kind]DefinitionCreator{
	reflect.Ptr:  newObjectDefinition,
	reflect.Func: newFunctionDefinition,
}

// 注册BeanDefinition创建器，使其能处理更多类型。
// 默认支持Pointer、Function
func RegisterBeanDefinitionCreator(kind reflect.Kind, creator DefinitionCreator) {
	if creator != nil {
		beanDefinitionCreators[kind] = creator
	}
}

func CreateBeanDefinition(o interface{}) (Definition, error) {
	t := reflect.TypeOf(o)

	creator, ok := beanDefinitionCreators[t.Kind()]
	if !ok || creator == nil {
		return nil, errors.New("Cannot handle this type: " + reflection.GetTypeName(t))
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
