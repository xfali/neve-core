// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package processor

import (
	"github.com/xfali/fig"
	"github.com/xfali/neve-core/bean"
)


type Processor interface {
	// 初始化对象处理器
	Init(conf fig.Properties, container bean.Container) error

	// 对象分类，判断对象是否实现某些接口，并进行相关归类。为了支持多协程处理，该方法应线程安全。
	// 注意：该方法建议只做归类，具体处理使用Process，不保证Processor的实现在此方法中做了相关处理。
	// 该方法在Bean Inject注入之后调用
	// return: bool 是否能够处理对象， error 处理是否有错误
	bean.Classifier

	// 对已分类对象做统一处理，注意如果存在耗时操作，请使用其他协程处理。
	// 该方法在Classify及BeanAfterSet后调用。
	// 成功返回nil，失败返回error
	Process() error

	// 资源回收相关操作
	bean.Disposable
}
