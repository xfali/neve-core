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

import (
	"time"
)

type ApplicationEvent interface {
	// 事件发生的时间
	OccurredTime() time.Time
}

type ApplicationEventPublisher interface {
	// 发送Context事件
	PublishEvent(e ApplicationEvent) error
}

type ApplicationEventListener interface {
	// 默认事件监听器接口
	// 监听器应尽快处理事件，耗时操作请使用协程
	OnApplicationEvent(e ApplicationEvent)
}

type ApplicationEventConsumerRegistry interface {
	// consumer: ApplicationEvent消费方法，类型func(ApplicationEvent)
	RegisterApplicationEventConsumer(consumer interface{}) error
}

type ApplicationEventConsumerListener interface {
	ApplicationEventListener
	ApplicationEventConsumerRegistry
}

type ApplicationEventConsumer interface {
	// 获得ApplicationEvent消费方法，类型func(ApplicationEvent)
	// 方法应尽快处理事件，耗时操作请使用协程
	RegisterConsumer(registry ApplicationEventConsumerRegistry) error
}

type ApplicationEventHandler interface {
	// 增加时间监听器
	// 监听器应尽快处理事件，耗时操作请使用协程
	AddListeners(listeners ...interface{})
}

type ApplicationEventProcessor interface {
	ApplicationEventPublisher
	ApplicationEventHandler

	// 同步通知事件
	// 不同于PublishEvent，NotifyEvent在Processor Close之后仍然能向Listener发送事件。
	NotifyEvent(e ApplicationEvent) error

	// 启动处理器，如有初始化操作必须定义在该方法
	Start() error

	// 停止处理，与Start方法对应，如有针对Start初始化的清理操作必须定义在该方法
	Close() error
}

type BaseApplicationEvent struct {
	timestamp time.Time
}

func (e *BaseApplicationEvent) ResetOccurredTime() {
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
