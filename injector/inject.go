// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

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

	// 从传入字串中解析注入名称和监听器
	ParseListener(v string) (name string, listeners []Listener)
}
