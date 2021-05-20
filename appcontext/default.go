// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package appcontext

import (
	"errors"
	"fmt"
	"github.com/xfali/fig"
	"github.com/xfali/neve-core/bean"
	"github.com/xfali/neve-core/injector"
	"github.com/xfali/neve-core/processor"
	"github.com/xfali/xlog"
	"reflect"
	"sync"
	"sync/atomic"
)

const (
	statusNone int32 = iota
	statusInitializing
	statusInitialized
)

const (
	defaultEventBufferSize = 4096
)

type Opt func(*defaultApplicationContext)

type defaultApplicationContext struct {
	config    fig.Properties
	logger    xlog.Logger
	container bean.Container
	injector  injector.Injector

	ctxAwares    []ApplicationContextAware
	ctxAwareLock sync.Mutex

	listeners    []ApplicationEventListener
	listenerLock sync.Mutex

	processors     []processor.Processor
	processorsLock sync.Mutex

	disableInject bool
	curState      int32

	eventBufSize int
	eventChan    chan ApplicationEvent
	stopChan     chan struct{}
}

func NewDefaultApplicationContext(opts ...Opt) *defaultApplicationContext {
	ret := &defaultApplicationContext{
		logger:    xlog.GetLogger(),
		container: bean.NewContainer(),

		curState:     statusNone,
		eventBufSize: defaultEventBufferSize,
		stopChan:     make(chan struct{}),
	}
	ret.injector = injector.New(injector.OptSetLogger(ret.logger))

	for _, opt := range opts {
		opt(ret)
	}

	ret.eventChan = make(chan ApplicationEvent, ret.eventBufSize)

	go ret.eventLoop()

	return ret
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

// set event channel buffer size
func OptSetEventBufferSize(size int) Opt {
	return func(context *defaultApplicationContext) {
		context.eventBufSize = size
	}
}

func (ctx *defaultApplicationContext) Init(config fig.Properties) (err error) {
	ctx.config = config
	ctx.disableInject = ctx.config.Get("neve.inject.disable", "false") == "true"
	return nil
}

func (ctx *defaultApplicationContext) Close() (err error) {
	close(ctx.stopChan)
	return
}

func (ctx *defaultApplicationContext) isInitializing() bool {
	return atomic.LoadInt32(&ctx.curState) == statusInitializing
}

func (ctx *defaultApplicationContext) RegisterBean(o interface{}, opts ...bean.RegisterOpt) error {
	return ctx.RegisterBeanByName("", o, opts...)
}

func (ctx *defaultApplicationContext) RegisterBeanByName(name string, o interface{}, opts ...bean.RegisterOpt) error {
	if ctx.isInitializing() {
		return errors.New("Initializing, cannot register new object. ")
	}

	if o == nil {
		return nil
	}
	var err error
	if name == "" {
		err = ctx.container.Register(o, opts...)
	} else {
		err = ctx.container.RegisterByName(name, o, opts...)
	}
	if err != nil {
		return err
	}

	ctx.processListener(o)

	if v, ok := o.(ApplicationContextAware); ok {
		ctx.addAware(v, true)
	}

	if v, ok := o.(processor.Processor); ok {
		err = ctx.addProcessor(v, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ctx *defaultApplicationContext) addListener(l ApplicationEventListener, withLock bool) {
	if !withLock {
		ctx.listeners = append(ctx.listeners, l)
		return
	}

	ctx.listenerLock.Lock()
	defer ctx.listenerLock.Unlock()

	ctx.listeners = append(ctx.listeners, l)
}

func (ctx *defaultApplicationContext) addAware(aware ApplicationContextAware, withLock bool) {
	if !withLock {
		ctx.ctxAwares = append(ctx.ctxAwares, aware)
		return
	}

	ctx.ctxAwareLock.Lock()
	defer ctx.ctxAwareLock.Unlock()
	ctx.ctxAwares = append(ctx.ctxAwares, aware)
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

func (ctx *defaultApplicationContext) AddProcessor(p processor.Processor) error {
	if p != nil {
		return ctx.addProcessor(p, true)
	}
	return errors.New("Processor is nil. ")
}

func (ctx *defaultApplicationContext) AddListeners(listeners ...interface{}) {
	for _, o := range listeners {
		ctx.processListener(o)
	}
}

func (ctx *defaultApplicationContext) processListener(o interface{}) {
	l, err := parseListener(o)
	if err != nil {
		//ctx.logger.Errorln(err)
	} else if l != nil {
		ctx.addListener(l, true)
	}
}

func (ctx *defaultApplicationContext) PublishEvent(e ApplicationEvent) error {
	if e == nil {
		return errors.New("event is nil. ")
	}
	select {
	case ctx.eventChan <- e:
		return nil
	default:
		return errors.New("event queue is full. ")
	}
}

func (ctx *defaultApplicationContext) Start() error {
	// 第一次初始化，注入所有对象
	if atomic.CompareAndSwapInt32(&ctx.curState, statusNone, statusInitializing) {
		// ApplicationContextAware Set.
		ctx.notifyAware()

		// Inject Beans
		if !ctx.disableInject {
			ctx.injectAll()
		}
		// Processor classify
		ctx.classifyBean()
		// Notify BeanAfterSet
		ctx.notifyBeanSet()
		// Processor process
		ctx.doProcess()
		// 初始化完成
		if !atomic.CompareAndSwapInt32(&ctx.curState, statusInitializing, statusInitialized) {
			ctx.logger.Fatal("Cannot be here!")
		}

		e := &ContextStartedEvent{}
		e.UpdateTime()
		e.ctx = ctx
		ctx.PublishEvent(e)
		return nil
	} else {
		return fmt.Errorf("Application Context Status error, current: %d . ", ctx.curState)
	}
}

func (ctx *defaultApplicationContext) notifyAware() {
	ctx.ctxAwareLock.Lock()
	defer ctx.ctxAwareLock.Unlock()

	for _, v := range ctx.ctxAwares {
		v.SetApplicationContext(ctx)
	}
}

func (ctx *defaultApplicationContext) classifyBean() {
	ctx.container.Scan(func(key string, value bean.Definition) bool {
		if value.IsObject() {
			// 必须先分类，由于ValueProcessor会在Classify将配置的属性值注入
			ctx.classifyOneBean(value.Interface())
		}
		return true
	})
}

func (ctx *defaultApplicationContext) classifyOneBean(o interface{}) {
	ctx.processorsLock.Lock()
	defer ctx.processorsLock.Unlock()

	for _, processor := range ctx.processors {
		_, err := processor.Classify(o)
		if err != nil {
			ctx.logger.Errorln(err)
		}
	}
}

func (ctx *defaultApplicationContext) notifyBeanSet() {
	ctx.container.Scan(func(key string, value bean.Definition) bool {
		err := value.AfterSet()
		if err != nil {
			ctx.logger.Errorln(err)
		}
		return true
	})
}

func (ctx *defaultApplicationContext) notifyEvent(e ApplicationEvent) {
	ctx.listenerLock.Lock()
	defer ctx.listenerLock.Unlock()

	for _, v := range ctx.listeners {
		v.OnApplicationEvent(e)
	}
}

func (ctx *defaultApplicationContext) doProcess() {
	ctx.processorsLock.Lock()
	defer ctx.processorsLock.Unlock()

	for _, processor := range ctx.processors {
		err := processor.Process()
		// processor error must return
		if err != nil {
			ctx.logger.Fatalln(err)
		}
	}
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

func (ctx *defaultApplicationContext) destroyBeans() {
	ctx.container.Scan(func(key string, value bean.Definition) bool {
		err := value.Destroy()
		if err != nil {
			ctx.logger.Errorln(err)
		}
		return true
	})
}

func (ctx *defaultApplicationContext) notifyClosed() {
	e := &ContextClosedEvent{}
	e.UpdateTime()
	e.ctx = ctx
	ctx.notifyEvent(e)
}

func (ctx *defaultApplicationContext) notifyStopped() {
	e := &ContextStoppedEvent{}
	e.UpdateTime()
	e.ctx = ctx
	ctx.notifyEvent(e)
}

func (ctx *defaultApplicationContext) eventLoop() {
	defer ctx.notifyClosed()
	defer ctx.destroyBeans()
	for {
		select {
		case <-ctx.stopChan:
			ctx.notifyStopped()
			return
		case e, ok := <-ctx.eventChan:
			if ok {
				ctx.notifyEvent(e)
			}
		}
	}
}

var eventType = reflect.TypeOf((*ApplicationEvent)(nil)).Elem()

type eventProcessor struct {
	et reflect.Type
	fv reflect.Value
}

func parseListener(o interface{}) (ApplicationEventListener, error) {
	if l, ok := o.(ApplicationEventListener); ok {
		return l, nil
	}

	t := reflect.TypeOf(o)
	if t.Kind() != reflect.Func {
		return nil, errors.New("Param is not a function. ")
	}

	if t.NumIn() != 1 {
		return nil, errors.New("Param is not match, expect func(ApplicationEvent). ")
	}

	et := t.In(0)
	if !et.Implements(eventType) {
		return nil, errors.New("Param is not match, function param must Implements ApplicationEvent. ")
	}

	return &eventProcessor{
		et: et,
		fv: reflect.ValueOf(o),
	}, nil
}

func (ep *eventProcessor) OnApplicationEvent(e ApplicationEvent) {
	t := reflect.TypeOf(e)
	//if t == ep.et {
	//	ep.fv.Call([]reflect.Value{reflect.ValueOf(e)})
	//}
	if t.AssignableTo(ep.et) {
		ep.fv.Call([]reflect.Value{reflect.ValueOf(e)})
	}
}
