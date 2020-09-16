// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package container

import (
	"errors"
	"github.com/xfali/neve/neve-utils/reflection"
	"reflect"
)

type BeanDefinition interface {
	// 类型
	Type() reflect.Type

	// 名称
	Name() string

	// 获得值
	Value() reflect.Value

	// 获得注册的bean对象
	Interface() interface{}

	// 是否注入的对象
	IsObject() bool
}

type objectDefinition struct {
	name string
	o    interface{}
	t    reflect.Type
}

func createBeanDefinition(o interface{}) (BeanDefinition, error) {
	t := reflect.TypeOf(o)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		if t.Kind() == reflect.Ptr {
			return nil, errors.New("o must be a Pointer but get Pointer's Pointer")
		}
		return newObjectDefinition(o), nil
	} else {
		err := verifyBeanFunction(t)
		if err != nil {
			return nil, err
		}
		return newFunctionDefinition(o), nil
	}

	return nil, errors.New("Cannot handle this type: " + reflection.GetTypeName(t))
}

func newObjectDefinition(o interface{}) *objectDefinition {
	t := reflect.TypeOf(o)
	return &objectDefinition{
		name: reflection.GetTypeName(t),
		o:    o,
		t:    t,
	}
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

var errType = reflect.TypeOf((*error)(nil)).Elem()

type functionDefinition struct {
	name string
	o    interface{}
	fn   reflect.Value
	t    reflect.Type
}

func verifyBeanFunction(ft reflect.Type) error {
	if ft.Kind() != reflect.Func {
		return errors.New("Param not function.")
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

func newFunctionDefinition(o interface{}) *functionDefinition {
	ft := reflect.TypeOf(o)
	ot := ft.Out(0)
	fn := reflect.ValueOf(o)
	return &functionDefinition{
		o:    o,
		name: reflection.GetTypeName(ot),
		fn:   fn,
		t:    ot,
	}
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
