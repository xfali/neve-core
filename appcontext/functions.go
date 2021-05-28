// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package appcontext

import (
	"github.com/xfali/neve-core/bean"
	"github.com/xfali/neve-core/injector"
)

type InjectFunctionRegistry interface {
	// function: 目标注入的方法，类型func(Type1, Type2...TypeN)
	// 可以通过参数names传递指定注入的对象名称。规则与tag:inject注入一致。如某参数选择自动注入则传字串""
	// 注意： names的长度要与functions的参数个数完全一致。
	RegisterInjectFunction(function interface{}, names ...string) error
}

// 如需要通过方法注入，则可实现该接口来注册方法，系统初始化会自动调用该方法进行注入。
type InjectFunction interface {
	// 获得ApplicationEvent消费方法，类型func(ApplicationEvent)
	// 方法应尽快处理事件，耗时操作请使用协程
	RegisterFunction(registry InjectFunctionRegistry) error
}

type InjectFunctionHandler interface {
	InjectFunctionRegistry

	// 设置injector
	SetInjector(injector injector.Injector)

	// 注入方法
	InjectAllFunctions(container bean.Container) error
}

type FunctionInjectInvoker interface {
	// 执行注入
	Invoke(injector injector.Injector, container bean.Container) error

	// 检查function是否符合类型要求
	ResolveFunction(injector injector.Injector, names []string, function interface{}) error
}
