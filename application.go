// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package neve

import (
	"github.com/xfali/fig"
	"github.com/xfali/neve-core/ctx"
	"github.com/xfali/neve-utils/log"
)

type Application interface {
	// 注册对象
	// 支持注册
	//  1、interface、struct指针，注册名称使用【类型名称】；
	//  2、struct/interface的构造函数 func() TYPE，注册名称使用【返回值的类型名称】。
	RegisterBean(o interface{}) error

	// 使用指定名称注册对象
	// 支持注册struct指针、struct/interface的构造函数 func() TYPE
	RegisterBeanByName(name string, o interface{}) error

	// 启动应用容器
	Run() error
}

type FileConfigApplication struct {
	ctx ctx.ApplicationContext
}

type Opt func(*FileConfigApplication)

func NewFileConfigApplication(configPath string, opts ...Opt) *FileConfigApplication {
	logger := log.GetLogger()
	prop, err := fig.LoadYamlFile(configPath)
	if err != nil {
		logger.Fatalln("load config file failed: ", err)
		return nil
	}
	ret := &FileConfigApplication{
		ctx: ctx.NewDefaultApplicationContext(prop),
	}

	for _, opt := range opts {
		opt(ret)
	}

	return ret
}

func (app *FileConfigApplication) RegisterBean(o interface{}) error {
	return app.ctx.RegisterBean(o)
}

func (app *FileConfigApplication) RegisterBeanByName(name string, o interface{}) error {
	return app.ctx.RegisterBeanByName(name, o)
}

func (app *FileConfigApplication) Run() error {
	app.ctx.NotifyListeners(ctx.ApplicationEventInitialized)
	return HandlerSignal(app.ctx.Close)
}

func OptSetApplicationContext(ctx ctx.ApplicationContext) Opt {
	return func(application *FileConfigApplication) {
		application.ctx = ctx
	}
}
