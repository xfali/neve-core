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

package injector

import (
	"github.com/xfali/neve-core/bean"
)

type InjectFunctionRegistry interface {
	// function: 目标注入的方法，类型func(Type1, Type2...TypeN)，该方法仅用于注入，不应做耗时及阻塞操作。
	// 可以通过参数names传递指定注入的对象名称。规则与tag:inject注入一致。如某参数选择自动注入则传字串""
	// 注意： names的长度要与function的参数个数完全一致。
	RegisterInjectFunction(function interface{}, names ...string) error
}

// 如需要通过方法注入，则可实现该接口来注册方法，系统初始化会自动调用该方法进行注入。
type InjectFunction interface {
	// 用于注册目标注入对象的方法
	// 注册的方法应尽快返回，不能进行耗时及阻塞操作。
	RegisterFunction(registry InjectFunctionRegistry) error
}

type InjectFunctionHandler interface {
	InjectFunctionRegistry

	// 设置injector
	SetInjector(injector Injector)

	// 注入方法
	InjectAllFunctions(container bean.Container) error
}

type FunctionInjectInvoker interface {
	// 执行注入的function名称
	FunctionName() string

	// 执行注入
	Invoke(injector Injector, container bean.Container, manager ListenerManager) error

	// 检查function是否符合类型要求
	ResolveFunction(injector Injector, names []string, function interface{}) error
}
