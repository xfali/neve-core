// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package processor

import (
	"github.com/xfali/fig"
	"github.com/xfali/neve/neve-core/container"
	"io"
)

type Processor interface {
	// 初始化对象处理器
	Init(conf fig.Properties, container container.Container) error

	// 对象分类，判断对象是否实现某些接口，并进行相关处理。
	// return: bool 是否能够处理对象， error 处理是否有错误
	Classify(o interface{}) (bool, error)

	// 统一处理
	Process() error

	// 关闭，资源回收相关操作
	io.Closer
}
