// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package bean

type Container interface {
	// 注册对象
	Register(o interface{}, opts ...RegisterOpt) error

	// 根据名称注册对象
	RegisterByName(name string, o interface{}, opts ...RegisterOpt) error

	// 根据名称获得对象
	// return：o：对象，ok：如果成功为true，否则为false
	Get(name string) (o interface{}, ok bool)

	// 根据类型获得对象，值设置到参数中
	// return：ok：如果成功为true，否则为false
	GetByType(o interface{}) bool

	// 获得对象定义
	// return：ok：如果成功为true，否则为false
	GetDefinition(name string) (d Definition, ok bool)

	// 添加对象定义
	// return：失败返回错误
	PutDefinition(name string, definition Definition) error

	// 遍历所有对象定义
	// f：key为注册的名称（可能为系统自动生成或手工指定）value为对象定义
	// 返回true则继续遍历，返回false停止遍历。
	Scan(f func(key string, value Definition) bool)
}
