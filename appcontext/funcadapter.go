// Copyright (C) 2020-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package appcontext

import "github.com/xfali/neve-core/injector"

type InjectFunctionRegistry = injector.InjectFunctionRegistry

// 如需要通过方法注入，则可实现该接口来注册方法，系统初始化会自动调用该方法进行注入。
//		用于注册目标注入对象的方法
//		注册的方法应尽快返回，不能进行耗时及阻塞操作。
//		RegisterFunction(registry appcontext.InjectFunctionRegistry) error
type InjectFunction = injector.InjectFunction
