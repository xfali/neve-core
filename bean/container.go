// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package bean

import (
	"errors"
	"github.com/xfali/neve-utils/reflection"
	"reflect"
	"sync"
)

type Container interface {
	// 注册对象
	Register(o interface{}) error

	// 根据名称注册对象
	RegisterByName(name string, o interface{}) error

	// 根据名称获得对象
	// return：o：对象，ok：如果成功为true，否则为false
	Get(name string) (o interface{}, ok bool)

	// 根据类型获得对象，值设置到参数中
	// return：ok：如果成功为true，否则为false
	GetByType(o interface{}) bool

	// 获得对象定义
	// return：ok：如果成功为true，否则为false
	GetDefinition(name string) (d Definition, ok bool)

	// 添加对象定义
	// return：失败返回错误
	PutDefinition(name string, definition Definition) error

	// 遍历所有对象定义
	// f：key为注册的名称（可能为系统自动生成或手工指定）value为对象定义
	// 返回true则继续遍历，返回false停止遍历。
	Scan(f func(key string, value Definition) bool)
}

type defaultContainer struct {
	objectPool sync.Map
}

func NewContainer() *defaultContainer {
	return &defaultContainer{}
}

func (c *defaultContainer) Register(o interface{}) error {
	return c.RegisterByName("", o)
}

func (c *defaultContainer) RegisterByName(name string, o interface{}) error {
	beanDefinition, err := CreateBeanDefinition(o)
	if err != nil {
		return err
	}
	if beanDefinition == nil {
		return errors.New("beanDefinition is nil. ")
	}

	if name == "" {
		name = reflection.GetObjectName(o)
		// func
		if name == "" {
			name = beanDefinition.Name()
		}
		if name == "" {
			return errors.New("Cannot get bean name. ")
		}
	}
	_, loaded := c.objectPool.LoadOrStore(name, beanDefinition)
	if loaded {
		return errors.New(name + " bean is exists. ")
	}
	return nil
}

func (c *defaultContainer) PutDefinition(name string, definition Definition) error {
	if definition == nil {
		return errors.New("Definition is nil. ")
	}
	_, loaded := c.objectPool.LoadOrStore(name, definition)
	if loaded {
		return errors.New(name + " bean is exists. ")
	}
	return nil
}

func (c *defaultContainer) GetDefinition(name string) (Definition, bool) {
	o, load := c.objectPool.Load(name)
	if load {
		return o.(Definition), load
	}
	return nil, false
}

func (c *defaultContainer) Get(name string) (interface{}, bool) {
	o, load := c.GetDefinition(name)
	if load {
		return o.Value().Interface(), load
	}
	return nil, false
}

func (c *defaultContainer) GetByType(o interface{}) bool {
	v := reflect.ValueOf(o)
	d, ok := c.GetDefinition(reflection.GetTypeName(v.Type()))
	if ok {
		v.Set(d.Value())
	}
	return false
}

func (c *defaultContainer) Scan(f func(key string, value Definition) bool) {
	c.objectPool.Range(func(key, value interface{}) bool {
		return f(key.(string), value.(Definition))
	})
}
