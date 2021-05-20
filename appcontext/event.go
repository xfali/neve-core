// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package appcontext

import (
	"time"
)

type ApplicationEvent interface {
	// 事件发生的时间
	OccurredTime() time.Time
}

type ApplicationContextAware interface {
	SetApplicationContext(cxt ApplicationContext)
}

type ApplicationEventPublisher interface {
	// 发送Context事件
	PublishEvent(e ApplicationEvent) error
}

type ApplicationEventHandler interface {
	// 增加时间监听器
	// 监听器应尽快处理事件，耗时操作请使用协程
	AddListeners(listeners ...interface{})
}

type ApplicationEventListener interface {
	// 默认事件监听器接口
	// 监听器应尽快处理事件，耗时操作请使用协程
	OnApplicationEvent(e ApplicationEvent)
}

type BaseApplicationEvent struct {
	timestamp time.Time
}

func (e *BaseApplicationEvent) UpdateTime() {
	e.timestamp = time.Now()
}

func (e *BaseApplicationEvent) OccurredTime() time.Time {
	return e.timestamp
}

type ApplicationContextEvent struct {
	BaseApplicationEvent
	ctx ApplicationContext
}

func (e *ApplicationContextEvent) GetContext() ApplicationContext {
	return e.ctx
}

// 服务启动后触发，Bean已经初始化完成，可以执行任意的业务逻辑
type ContextStartedEvent struct {
	ApplicationContextEvent
}

// 服务停止后触发，应尽快做清理工作
type ContextStoppedEvent struct {
	ApplicationContextEvent
}

// 已到达ApplicationContext生命周期末端，应用即将退出
type ContextClosedEvent struct {
	ApplicationContextEvent
}