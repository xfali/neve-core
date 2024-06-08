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

package appcontext

import (
	"context"
	"errors"
	"fmt"
	"github.com/xfali/fig"
	"github.com/xfali/neve-core/bean"
	"github.com/xfali/neve-core/injector"
	"github.com/xfali/neve-core/processor"
	"github.com/xfali/neve-core/version"
	"github.com/xfali/xlog"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	statusNone int32 = iota
	statusInitializing
	statusInitialized
)

type Opt func(*defaultApplicationContext)

type defaultApplicationContext struct {
	config       fig.Properties
	logger       xlog.Logger
	container    bean.Container
	injector     injector.Injector
	funcHandler  injector.InjectFunctionHandler
	injectLicMgr injector.ListenerManager
	eventProc    ApplicationEventProcessor

	ctxAwares    []ApplicationContextAware
	ctxAwareLock sync.Mutex

	processors     []processor.Processor
	processorsLock sync.Mutex

	appName       string
	disableInject bool
	disableEvent  bool
	curState      int32

	closeOnce sync.Once
}

func NewDefaultApplicationContext(opts ...Opt) *defaultApplicationContext {
	ret := &defaultApplicationContext{
		logger:    xlog.GetLogger(),
		container: bean.NewContainer(),
		eventProc: NewEventProcessor(),

		curState: statusNone,
	}
	ret.injectLicMgr = injector.NewListenerManager(ret.logger)
	ret.injector = injector.New(injector.OptSetLogger(ret.logger), injector.OptSetListenerManager(ret.injectLicMgr))
	ret.funcHandler = injector.NewDefaultInjectFunctionHandler(ret.logger, ret.injectLicMgr)

	for _, opt := range opts {
		opt(ret)
	}

	if setter, ok := ret.injector.(injector.ListenerManagerSetter); ok {
		setter.SetListenerManager(ret.injectLicMgr)
	}
	if setter, ok := ret.funcHandler.(injector.ListenerManagerSetter); ok {
		setter.SetListenerManager(ret.injectLicMgr)
	}
	ret.funcHandler.SetInjector(ret.injector)

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

func OptSetInjectListenerManager(manager injector.ListenerManager) Opt {
	return func(context *defaultApplicationContext) {
		context.injectLicMgr = manager
	}
}

func OptSetInjector(injector injector.Injector) Opt {
	return func(context *defaultApplicationContext) {
		context.injector = injector
	}
}

func OptSetInjectFunctionHandler(handler injector.InjectFunctionHandler) Opt {
	return func(context *defaultApplicationContext) {
		context.funcHandler = handler
	}
}

func OptSetEventProcessor(proc ApplicationEventProcessor) Opt {
	return func(context *defaultApplicationContext) {
		context.eventProc = proc
	}
}

func OptDisableEvent() Opt {
	return func(context *defaultApplicationContext) {
		context.disableEvent = true
	}
}

func (ctx *defaultApplicationContext) Init(config fig.Properties) (err error) {
	ctx.config = config
	ctx.appName = ctx.config.Get("neve.application.name", "Neve Application")
	ctx.disableInject = ctx.config.Get("neve.inject.disable", "false") == "true"

	event := ctx.config.Get("neve.application.eventMode", "on")
	event = strings.ToLower(event)
	if !ctx.disableEvent {
		ctx.disableEvent = event == "off" || event == "false"
	}

	if ctx.disableEvent && ctx.eventProc != nil {
		ctx.eventProc = NewDisableEventProcessor()
	}
	// Register ApplicationEventPublisher
	ctx.container.Register(ctx.eventProc.(ApplicationEventPublisher))

	return ctx.eventProc.Start()
}

func (ctx *defaultApplicationContext) GetApplicationName() string {
	return ctx.appName
}

func (ctx *defaultApplicationContext) Close() (err error) {
	ctx.closeOnce.Do(func() {
		err = ctx.eventProc.Close()
		if err != nil {
			ctx.logger.Errorln(err)
		}
		ctx.notifyStopped()
		ctx.destroyBeans()
		ctx.notifyClosed()
	})

	return nil
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
	o, err = injector.WrapBean(o, ctx.container, ctx.injector, ctx.injectLicMgr)
	if err != nil {
		return err
	}

	if name == "" {
		err = ctx.container.Register(o, opts...)
	} else {
		err = ctx.container.RegisterByName(name, o, opts...)
	}
	if err != nil {
		return err
	}

	if !ctx.disableEvent {
		ctx.eventProc.AddListeners(o)
	}

	err = ctx.classifyInjectFunction(o)
	if err != nil {
		return err
	}

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

func (ctx *defaultApplicationContext) classifyInjectFunction(o interface{}) error {
	if v, ok := o.(injector.InjectFunction); ok {
		return v.RegisterFunction(ctx.funcHandler)
	}
	return nil
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
	ctx.eventProc.AddListeners(listeners...)
}

func (ctx *defaultApplicationContext) PublishEvent(e ApplicationEvent) error {
	return ctx.eventProc.PublishEvent(e)
}

func (ctx *defaultApplicationContext) PostEvent(context context.Context, e ApplicationEvent) error {
	return ctx.eventProc.PostEvent(context, e)
}

func (ctx *defaultApplicationContext) SendEvent(e ApplicationEvent) error {
	return ctx.eventProc.SendEvent(e)
}

func (ctx *defaultApplicationContext) Start() error {
	ctx.printCtxInfo()
	// 第一次初始化，注入所有对象
	if atomic.CompareAndSwapInt32(&ctx.curState, statusNone, statusInitializing) {
		// ApplicationContextAware Set.
		ctx.notifyAware()

		// Inject Beans
		ctx.injectAll()
		// Processor classify
		ctx.classifyBean()
		// call and inject all functions
		ctx.doFunctionInject()
		// Notify BeanAfterSet
		ctx.notifyBeanSet()
		// Processor process
		ctx.doProcess()

		// 初始化完成
		if !atomic.CompareAndSwapInt32(&ctx.curState, statusInitializing, statusInitialized) {
			ctx.logger.Fatal("Cannot be here!")
		}

		ctx.notifyStarted()
		return nil
	} else {
		return fmt.Errorf("Application Context Status error, current: %d . ", ctx.curState)
	}
}

func (ctx *defaultApplicationContext) printCtxInfo() {
	path := ctx.config.Get("neve.application.banner", "")
	mode := ctx.config.Get("neve.application.bannerMode", "")
	mode = strings.ToLower(mode)
	printNeveInfo(version.NeveVersion, path, mode != "off" && mode != "false")
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
		//if value.IsObject() {
		// 必须先分类，由于ValueProcessor会在Classify将配置的属性值注入
		ctx.classifyOneBean(value)
		//}
		return true
	})
}

func (ctx *defaultApplicationContext) classifyOneBean(o bean.Definition) {
	ctx.processorsLock.Lock()
	defer ctx.processorsLock.Unlock()

	for _, processor := range ctx.processors {
		_, err := o.Classify(processor)
		//_, err := processor.Classify(o)
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

func (ctx *defaultApplicationContext) doFunctionInject() {
	if ctx.disableInject {
		return
	}
	err := ctx.funcHandler.InjectAllFunctions(ctx.container)
	if err != nil {
		ctx.logger.Errorln(err)
	}
}

func (ctx *defaultApplicationContext) injectAll() {
	if ctx.disableInject {
		return
	}
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

func (ctx *defaultApplicationContext) notifyStarted() {
	if ctx.disableEvent {
		return
	}
	_ = ctx.PublishEvent(NewContextStartedEvent(ctx))
}

func (ctx *defaultApplicationContext) notifyClosed() {
	if ctx.disableEvent {
		return
	}
	_ = ctx.eventProc.SendEvent(NewContextClosedEvent(ctx))
}

func (ctx *defaultApplicationContext) notifyStopped() {
	if ctx.disableEvent {
		return
	}
	_ = ctx.eventProc.SendEvent(NewContextStoppedEvent(ctx))
}
