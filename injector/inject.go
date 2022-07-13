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
	"reflect"
)

type Injector interface {
	// 是否可以注入，是返回true，否则返回false
	// 该方法为了保证效率可能仅为初步类型筛查:
	//    返回false说明该对象无法注入。
	//    返回true说明该对象可能注入成功，但以Inject实际注入时的返回为准。
	CanInject(o interface{}) bool

	// 从对象容器中注入对象到参数o
	// return: 成功返回nil，否则返回错误原因
	Inject(container bean.Container, o interface{}) error

	// 判断目标类型是否可以注入
	// 该方法为了保证效率可能仅为初步类型筛查:
	//    返回false说明该类型无法注入。
	//    返回true说明该类型可能注入成功，但以InjectValue实际注入时的返回为准。
	CanInjectType(t reflect.Type) bool

	// 从对象容器中注入对象到value
	// return: 成功返回nil，否则返回错误原因
	InjectValue(c bean.Container, name string, v reflect.Value) error
}

type Actuator func(c bean.Container, name string, v reflect.Value) error

// 注入监听器
type Listener interface {
	// 当注入失败时回调
	OnInjectFailed(err error)
}

// 监听管理器
type ListenerManager interface {
	// 添加监听器
	AddListener(name string, listener Listener)

	// 从传入字串中解析注入名称和匹配监听器
	ParseListener(v string) (name string, listeners []Listener)
}
