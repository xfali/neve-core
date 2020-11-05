// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package appcontext

import (
	"errors"
	"github.com/xfali/fig"
	"github.com/xfali/neve-core/bean"
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
	// 初始化context
	Init(config fig.Properties) error

	// 注册对象
	RegisterBean(o interface{}) error

	// 使用指定名称注册对象
	RegisterBeanByName(name string, o interface{}) error

	// 根据名称获得对象，如果容器中包含该对象，则返回对象和true否则返回nil和false
	GetBean(name string) (interface{}, bool)

	// 从容器注入对象，如果容器中包含该对象，返回true否则返回false
	GetBeanByType(o interface{}) bool

	// 增加对象处理器，用于对对象进行分类和处理
	AddProcessor(processor.Processor)

	// 增加Context的监听器
	AddListener(ApplicationContextListener)

	// 通知Context事件
	NotifyListeners(ApplicationEvent)

	// 关闭，用于资源回收
	Close() error
}

type ApplicationContextListener interface {
	OnRefresh(ctx ApplicationContext)
	OnEvent(e ApplicationEvent, ctx ApplicationContext)
}

type Opt func(*defaultApplicationContext)

type defaultApplicationContext struct {
	config    fig.Properties
	logger    xlog.Logger
	container bean.Container
	injector  injector.Injector

	listeners    []ApplicationContextListener
	listenerLock sync.Mutex

	processors     []processor.Processor
	processorsLock sync.Mutex

	curState int32
}

func NewDefaultApplicationContext(opts ...Opt) *defaultApplicationContext {
	ret := &defaultApplicationContext{
		logger:    xlog.GetLogger(),
		container: bean.NewContainer(),

		curState: statusNone,
	}
	ret.injector = injector.New(injector.OptSetLogger(ret.logger))

	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func (ctx *defaultApplicationContext) Init(config fig.Properties) (err error) {
	ctx.config = config
	return nil
}

func (ctx *defaultApplicationContext) Close() (err error) {
	ctx.container.Scan(func(key string, value bean.Definition) bool {
		err = value.Destroy()
		if err != nil {
			ctx.logger.Errorln(err)
		}
		return true
	})

	return
}

func (ctx *defaultApplicationContext) isInitializing() bool {
	return atomic.LoadInt32(&ctx.curState) == statusInitializing
}

func (ctx *defaultApplicationContext) RegisterBean(o interface{}) error {
	return ctx.RegisterBeanByName("", o)
}

func (ctx *defaultApplicationContext) RegisterBeanByName(name string, o interface{}) error {
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

	if v, ok := o.(ApplicationContextListener); ok {
		ctx.addListener(v, true)
	}

	if v, ok := o.(processor.Processor); ok {
		ctx.addProcessor(v, true)
	}

	ctx.listenerLock.Lock()
	defer ctx.listenerLock.Unlock()

	for _, v := range ctx.listeners {
		v.OnRefresh(ctx)
	}
	return nil
}

func (ctx *defaultApplicationContext) addListener(l ApplicationContextListener, withLock bool) {
	if !withLock {
		ctx.listeners = append(ctx.listeners, l)
		return
	}

	ctx.listenerLock.Lock()
	defer ctx.listenerLock.Unlock()

	ctx.listeners = append(ctx.listeners, l)
}

func (ctx *defaultApplicationContext) addProcessor(p processor.Processor, withLock bool) error {
	if !withLock {
		ctx.processors = append(ctx.processors, p)
		return p.Init(ctx.config, ctx.container)
	}

	ctx.processorsLock.Lock()
	defer ctx.processorsLock.Unlock()

	ctx.processors = append(ctx.processors, p)
	return p.Init(ctx.config, ctx.container)
}

func (ctx *defaultApplicationContext) GetBean(name string) (interface{}, bool) {
	return ctx.container.Get(name)
}

func (ctx *defaultApplicationContext) GetBeanByType(o interface{}) bool {
	return ctx.container.GetByType(o)
}

func (ctx *defaultApplicationContext) AddProcessor(p processor.Processor) {
	ctx.processors = append(ctx.processors, p)
}

func (ctx *defaultApplicationContext) AddListener(l ApplicationContextListener) {
	ctx.addListener(l, true)
}

func (ctx *defaultApplicationContext) NotifyListeners(e ApplicationEvent) {
	if ApplicationEventInitialized == e {
		// 第一次初始化，注入所有对象
		if atomic.CompareAndSwapInt32(&ctx.curState, statusNone, statusInitializing) {
			if ctx.config.Get("neve.inject.disable", "false") != "true" {
				ctx.injectAll()
			}
			err := ctx.processBean()
			if err != nil {
				ctx.logger.Fatalln(err)
			}
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

func (ctx *defaultApplicationContext) processBean() error {
	ctx.container.Scan(func(key string, value bean.Definition) bool {
		if value.IsObject() {
			// 必须先分类，由于ValueProcessor会在Classify将配置的属性值注入
			for _, processor := range ctx.processors {
				_, err := processor.Classify(value.Interface())
				if err != nil {
					ctx.logger.Errorln(err)
				}
			}

			err := value.AfterSet()
			if err != nil {
				ctx.logger.Errorln(err)
			}
		}
		return true
	})

	for _, processor := range ctx.processors {
		err := processor.Process()
		// processor error must return
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctx *defaultApplicationContext) injectAll() {
	ctx.container.Scan(func(key string, value bean.Definition) bool {
		if value.IsObject() {
			err := ctx.injector.Inject(ctx.container, value.Interface())
			if err != nil {
				ctx.logger.Errorln("Inject failed: ", err)
			}
		}
		return true
	})
}

func OptSetContainer(container bean.Container) Opt {
	return func(context *defaultApplicationContext) {
		context.container = container
	}
}

func OptSetLogger(logger xlog.Logger) Opt {
	return func(context *defaultApplicationContext) {
		context.logger = logger
	}
}

func OptSetInjector(injector injector.Injector) Opt {
	return func(context *defaultApplicationContext) {
		context.injector = injector
	}
}
