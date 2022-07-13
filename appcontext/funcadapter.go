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

import "github.com/xfali/neve-core/injector"

type InjectFunctionRegistry = injector.InjectFunctionRegistry

// 如需要通过方法注入，则可实现该接口来注册方法，系统初始化会自动调用该方法进行注入。
//		用于注册目标注入对象的方法
//		注册的方法应尽快返回，不能进行耗时及阻塞操作。
//		RegisterFunction(registry appcontext.InjectFunctionRegistry) error
type InjectFunction = injector.InjectFunction
