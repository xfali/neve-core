// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package ctx

import (
	"errors"
	"github.com/xfali/neve-core/container"
	"github.com/xfali/neve-core/injector"
	"github.com/xfali/neve-core/processor"
	"github.com/xfali/xlog"
	"sync"
	"sync/atomic"
)

type ApplicationEvent int

const (
	ApplicationEventNone ApplicationEvent = iota
	ApplicationEventInitialized
)

const (
	statusNone int32 = iota
	statusInitializing
	statusInitialized
)

type ApplicationContext interface {
	RegisterBean(o interface{}) error
	RegisterBeanByName(name string, o interface{}) error

	GetBean(name string) (interface{}, bool)
	GetBeanByType(o interface{}) bool

	AddProcessor(processor.Processor)

	AddListener(ApplicationContextListener)
	NotifyListeners(ApplicationEvent)

	Close() error
}

type ApplicationContextListener interface {
	OnRefresh(ctx ApplicationContext)
	OnEvent(e ApplicationEvent, ctx ApplicationContext)
}

type Opt func(*DefaultApplicationContext)

type DefaultApplicationContext struct {
	logger    xlog.Logger
	container container.Container
	injector  injector.Injector

	listeners    []ApplicationContextListener
	listenerLock sync.Mutex

	processors []processor.Processor

	curState int32
}

func NewDefaultApplicationContext(opts ...Opt) *DefaultApplicationContext {
	ret := &DefaultApplicationContext{
		logger:    xlog.GetLogger(),
		container: container.New(),

		curState: statusNone,
	}
	ret.injector = injector.New(injector.OptSetLogger(ret.logger))

	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func (ctx *DefaultApplicationContext) Close() (err error) {
	for _, processor := range ctx.processors {
		err = processor.Close()
		if err != nil {
			ctx.logger.Errorln(err)
		}
	}
	return
}

func (ctx *DefaultApplicationContext) isInitializing() bool {
	return atomic.LoadInt32(&ctx.curState) == statusInitializing
}

func (ctx *DefaultApplicationContext) RegisterBean(o interface{}) error {
	return ctx.RegisterBeanByName("", o)
}

func (ctx *DefaultApplicationContext) RegisterBeanByName(name string, o interface{}) error {
	if ctx.isInitializing() {
		return errors.New("Initializing, cannot register new object. ")
	}

	if o == nil {
		return nil
	}
	var err error
	if name == "" {
		err = ctx.container.Register(o)
	} else {
		err = ctx.container.RegisterByName(name, o)
	}

	if err != nil {
		return err
	}

	for _, processor := range ctx.processors {
		_, err := processor.Classify(o)
		if err != nil {
			ctx.logger.Errorln(err)
		}
	}

	ctx.listenerLock.Lock()
	defer ctx.listenerLock.Unlock()

	for _, v := range ctx.listeners {
		v.OnRefresh(ctx)
	}

	switch o.(type) {
	case ApplicationContextListener:
		ctx.listeners = append(ctx.listeners, o.(ApplicationContextListener))
	}
	return nil
}

func (ctx *DefaultApplicationContext) GetBean(name string) (interface{}, bool) {
	return ctx.container.Get(name)
}

func (ctx *DefaultApplicationContext) GetBeanByType(o interface{}) bool {
	return ctx.container.GetByType(o)
}

func (ctx *DefaultApplicationContext) AddProcessor(p processor.Processor) {
	ctx.processors = append(ctx.processors, p)
}

func (ctx *DefaultApplicationContext) AddListener(l ApplicationContextListener) {
	ctx.listenerLock.Lock()
	defer ctx.listenerLock.Unlock()

	ctx.listeners = append(ctx.listeners, l)
}

func (ctx *DefaultApplicationContext) NotifyListeners(e ApplicationEvent) {
	if ApplicationEventInitialized == e {
		for _, processor := range ctx.processors {
			processor.Process()
		}
		// 第一次初始化，注入所有对象
		if atomic.CompareAndSwapInt32(&ctx.curState, statusNone, statusInitializing) {
			ctx.injectAll()
			// 初始化完成
			if !atomic.CompareAndSwapInt32(&ctx.curState, statusInitializing, statusInitialized) {
				ctx.logger.Fatal("Cannot be here!")
			}
		}
	}

	ctx.listenerLock.Lock()
	defer ctx.listenerLock.Unlock()
	for _, v := range ctx.listeners {
		v.OnEvent(e, ctx)
	}
}

func (ctx *DefaultApplicationContext) injectAll() {
	ctx.container.Scan(func(key string, value container.BeanDefinition) bool {
		if value.IsObject() {
			err := ctx.injector.Inject(ctx.container, value.Interface())
			if err != nil {
				ctx.logger.Errorln("Inject failed: ", err)
			}
		}
		return true
	})
}

func OptSetContainer(container container.Container) Opt {
	return func(context *DefaultApplicationContext) {
		context.container = container
	}
}

func OptSetLogger(logger xlog.Logger) Opt {
	return func(context *DefaultApplicationContext) {
		context.logger = logger
	}
}

func OptSetInjector(injector injector.Injector) Opt {
	return func(context *DefaultApplicationContext) {
		context.injector = injector
	}
}
