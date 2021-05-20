// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package appcontext

import (
	"github.com/xfali/fig"
	"github.com/xfali/neve-core/bean"
	"github.com/xfali/neve-core/processor"
)

type ApplicationContext interface {
	// 初始化context
	Init(config fig.Properties) error

	// 获得应用名称
	GetApplicationName() string

	// 注册对象
	// opts添加bean注册的配置，详情查看bean.RegisterOpt
	RegisterBean(o interface{}, opts ...bean.RegisterOpt) error

	// 使用指定名称注册对象
	// opts添加bean注册的配置，详情查看bean.RegisterOpt
	RegisterBeanByName(name string, o interface{}, opts ...bean.RegisterOpt) error

	// 根据名称获得对象，如果容器中包含该对象，则返回对象和true否则返回nil和false
	GetBean(name string) (interface{}, bool)

	// 从容器注入对象，如果容器中包含该对象，返回true否则返回false
	GetBeanByType(o interface{}) bool

	// 增加对象处理器，用于对对象进行分类和处理
	AddProcessor(processor.Processor) error

	// 启动应用
	Start() error

	// 关闭，用于资源回收
	Close() error

	ApplicationEventPublisher

	ApplicationEventHandler
}

type ApplicationContextAware interface {
	// 装配ApplicationContext
	// 在bean未被注入和初始化之前调用
	SetApplicationContext(ctx ApplicationContext)
}
