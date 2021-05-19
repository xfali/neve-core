// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package bean

import (
	"errors"
	"github.com/xfali/goutils/container/skiplist"
	"github.com/xfali/neve-utils/reflection"
	"reflect"
	"sync"
)

const (
	defaultPoolSize = 128
	defaultOrder    = 0
)

func NewContainer() *defaultContainer {
	return &defaultContainer{
		objectPool: newPool(defaultPoolSize),
	}
}

type elem struct {
	def   Definition
	order int
}

func newElem(opts ...RegisterOpt) *elem {
	ret := &elem{
		order: defaultOrder,
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func compareElem(a, b interface{}) int {
	return a.(*elem).order - b.(*elem).order
}

func (e *elem) Set(key string, value interface{}) {
	if key == KeySetOrder {
		e.order = value.(int)
	}
}

type pool struct {
	l *skiplist.SkipList
	m map[string]*elem

	locker sync.Mutex
}

func newPool(initSize int) *pool {
	return &pool{
		m: make(map[string]*elem, initSize),
		l: skiplist.New(skiplist.SetKeyCompareFunc(skiplist.CompareInt)),
	}
}

func (p *pool) keys() []string {
	p.locker.Lock()
	defer p.locker.Unlock()

	if p.l.Len() == 0 {
		return nil
	}
	ret := make([]string, 0, len(p.m))
	for x := p.l.First(); x != nil; x = x.Next() {
		ret = append(ret, x.Value().([]string)...)
	}
	return ret
}

func (p *pool) loadOrStore(name string, elem *elem) (*elem, bool) {
	p.locker.Lock()
	defer p.locker.Unlock()

	if v, ok := p.m[name]; ok {
		return v, true
	} else {
		keys := p.l.Get(elem.order)
		if keys == nil {
			keys = []string{name}
		} else {
			keys = append(keys.([]string), name)
		}
		p.l.Set(elem.order, keys)
		p.m[name] = elem
		return elem, false
	}
}

func (p *pool) load(name string) (*elem, bool) {
	p.locker.Lock()
	defer p.locker.Unlock()

	if v, ok := p.m[name]; ok {
		return v, true
	} else {
		return nil, false
	}
}

type defaultContainer struct {
	objectPool *pool
}

func (c *defaultContainer) Register(o interface{}, opts ...RegisterOpt) error {
	return c.RegisterByName("", o, opts...)
}

func (c *defaultContainer) RegisterByName(name string, o interface{}, opts ...RegisterOpt) error {
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

	elem := newElem(opts...)
	elem.def = beanDefinition
	_, loaded := c.objectPool.loadOrStore(name, elem)
	if loaded {
		return errors.New(name + " bean is exists. ")
	}
	return nil
}

func (c *defaultContainer) PutDefinition(name string, definition Definition) error {
	if definition == nil {
		return errors.New("Definition is nil. ")
	}
	elem := newElem()
	elem.def = definition
	_, loaded := c.objectPool.loadOrStore(name, elem)
	if loaded {
		return errors.New(name + " bean is exists. ")
	}
	return nil
}

func (c *defaultContainer) GetDefinition(name string) (Definition, bool) {
	o, load := c.objectPool.load(name)
	if load {
		return o.def, load
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
	keys := c.objectPool.keys()
	for _, k := range keys {
		if v, ok := c.objectPool.load(k); ok {
			if !f(k, v.def) {
				break
			}
		}
	}
}