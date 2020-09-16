// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package container

import (
	"errors"
	"github.com/xfali/neve-utils/reflection"
	"reflect"
	"sync"
)

type Container interface {
	Register(o interface{}) error
	RegisterByName(name string, o interface{}) error

	Get(name string) (BeanDefinition, bool)
	GetByType(o interface{}) bool

	Scan(func(key string, value BeanDefinition) bool)
}

type DefaultContainer struct {
	objectPool sync.Map
}

func New() *DefaultContainer {
	return &DefaultContainer{}
}

func (c *DefaultContainer) Register(o interface{}) error {
	return c.RegisterByName("", o)
}

func (c *DefaultContainer) RegisterByName(name string, o interface{}) error {
	beanDefinition, err := createBeanDefinition(o)
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
	}
	_, loaded := c.objectPool.LoadOrStore(name, beanDefinition)
	if loaded {
		return errors.New(name + " bean is exists. ")
	}
	return nil
}

func (c *DefaultContainer) Get(name string) (BeanDefinition, bool) {
	o, load := c.objectPool.Load(name)
	if load {
		return o.(BeanDefinition), load
	}
	return nil, false
}

func (c *DefaultContainer) GetByType(o interface{}) bool {
	v := reflect.ValueOf(o)
	d, ok := c.Get(reflection.GetTypeName(v.Type()))
	if ok {
		v.Set(d.Value())
	}
	return false
}

func (c *DefaultContainer) Scan(f func(key string, value BeanDefinition) bool) {
	c.objectPool.Range(func(key, value interface{}) bool {
		return f(key.(string), value.(BeanDefinition))
	})
}
