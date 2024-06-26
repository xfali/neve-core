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

package neve

import (
	"context"
	"github.com/xfali/fig"
	"github.com/xfali/neve-core/appcontext"
	"github.com/xfali/neve-core/application"
	"github.com/xfali/neve-core/bean"
	"github.com/xfali/neve-core/injector"
	"github.com/xfali/xlog"
	"strconv"
	"time"
)

type Application interface {
	// RegisterBean 注册对象
	// 支持注册
	//  1、interface、struct指针，注册名称使用【类型名称】；
	//  2、struct/interface的构造函数 func() TYPE，注册名称使用【返回值的类型名称】。
	// opts添加bean注册的配置，详情查看bean.RegisterOpt
	RegisterBean(o interface{}, opts ...RegisterOpt) error

	// RegisterBeanByName 使用指定名称注册对象
	// 支持注册
	//  1、interface、struct指针，注册名称使用【类型名称】；
	//  2、struct/interface的构造函数 func() TYPE，注册名称使用【返回值的类型名称】。
	// opts添加bean注册的配置，详情查看bean.RegisterOpt
	RegisterBeanByName(name string, o interface{}, opts ...RegisterOpt) error

	// AddListeners 增加时间监听器
	// 监听器应尽快处理事件，耗时操作请使用协程
	AddListeners(listeners ...interface{})

	// Run 启动应用容器
	Run() error

	// RunWithContext 启动应用容器
	// 参数ctx: 应用ctx，如果ctx Done则退出
	RunWithContext(ctx context.Context) error

	// Stop 强制退出
	Stop()
}

const (
	QuitSleepTime = 3 * time.Second
)

type RegisterOpt = bean.RegisterOpt

type FileConfigApplication struct {
	ctx    appcontext.ApplicationContext
	waiter application.SignalWaiter
	logger xlog.Logger

	quitSleepTime time.Duration
}

type Opt func(*FileConfigApplication)

func NewApplication(prop fig.Properties, opts ...Opt) *FileConfigApplication {
	if prop == nil {
		xlog.Errorln("Properties cannot be nil. ")
		return nil
	}
	ret := &FileConfigApplication{
		ctx:           appcontext.NewDefaultApplicationContext(),
		logger:        xlog.GetLogger(),
		quitSleepTime: QuitSleepTime,
	}

	ret.waiter = application.NewSignalWaiter(application.SignalWaiterOpts.SetLogger(ret.logger))
	for _, opt := range opts {
		opt(ret)
	}

	err := ret.ctx.Init(prop)
	if err != nil {
		ret.logger.Fatalln(err)
		return nil
	}

	sleepTime := prop.Get("neve.application.quit.sleepSec", "")
	if sleepTime != "" {
		if t, pErr := strconv.Atoi(sleepTime); pErr == nil {
			if t >= 0 {
				ret.quitSleepTime = time.Duration(t) * time.Second
			}
		}
	}

	return ret
}

func NewFileConfigApplication(configPath string, opts ...Opt) *FileConfigApplication {
	// Disable fig's log
	fig.SetLog(func(format string, o ...interface{}) {})
	prop, err := fig.LoadYamlFile(configPath)
	if err != nil {
		xlog.Errorln("load config file failed: ", err)
		return nil
	}
	return NewApplication(prop, opts...)
}

func (app *FileConfigApplication) RegisterBean(o interface{}, opts ...RegisterOpt) error {
	return app.ctx.RegisterBean(o, opts...)
}

func (app *FileConfigApplication) RegisterBeanByName(name string, o interface{}, opts ...RegisterOpt) error {
	return app.ctx.RegisterBeanByName(name, o, opts...)
}

func (app *FileConfigApplication) AddListeners(listeners ...interface{}) {
	app.ctx.AddListeners(listeners...)
}

func (app *FileConfigApplication) Run() error {
	return app.RunWithContext(context.Background())
}

func (app *FileConfigApplication) RunWithContext(ctx context.Context) (err error) {
	err = app.ctx.Start()
	if err != nil {
		return err
	}
	defer func(pErr *error) {
		*pErr = Quit(app.logger, app.quitSleepTime, app.ctx.Close)
		if *pErr != nil {
			app.logger.Errorln(*pErr)
		}
	}(&err)
	_ = app.waiter.Wait(ctx)
	return
}

func (app *FileConfigApplication) Stop() {
	app.waiter.Stop()
}

func OptSetApplicationContext(ctx appcontext.ApplicationContext) Opt {
	return func(application *FileConfigApplication) {
		application.ctx = ctx
	}
}

func OptSetLogger(logger xlog.Logger) Opt {
	return func(application *FileConfigApplication) {
		application.logger = logger
	}
}

func OptSetSignalWaiter(waiter application.SignalWaiter) Opt {
	return func(application *FileConfigApplication) {
		application.waiter = waiter
	}
}

func OptSetQuitSleepTime(t time.Duration) Opt {
	return func(application *FileConfigApplication) {
		application.quitSleepTime = t
	}
}

func OptSetInjectTagName(name string) Opt {
	return func(application *FileConfigApplication) {
		application.ctx = appcontext.NewDefaultApplicationContext(
			appcontext.OptSetInjector(injector.New(
				injector.OptSetLogger(application.logger),
				injector.OptSetInjectTagName(name))))
	}
}

func SetOrder(order int) RegisterOpt {
	return bean.SetOrder(order)
}
