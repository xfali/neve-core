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
	"context"
	"time"
)

type ApplicationEvent interface {
	// 事件发生的时间
	OccurredTime() time.Time
}

type EventContextHolder interface {
	// 获得事件的context
	GetEventContext() context.Context
}

type ContextEvent interface {
	ApplicationEvent
	EventContextHolder
}

type ApplicationEventPublisher interface {
	// PublishEvent 发送Application事件（异步处理）
	// 该方法不会阻塞，如果事件队列已满则直接返回错误
	// e: Application事件
	PublishEvent(e ApplicationEvent) error

	// PostEvent 发送Application事件（异步处理）
	// 如果事件队列已满则会阻塞，直至事件成功加入队列或者ctx被cancel
	// ctx: 事件处理的context，不可为nil
	// e: Application事件
	PostEvent(ctx context.Context, e ApplicationEvent) error

	// SendEvent 发送Application事件（同步处理）
	// 该方法会阻塞，直至事件全部发送完成
	// e: Application事件
	SendEvent(e ApplicationEvent) error
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

	// Deprecated: 请使用SendEvent(ctx context.Context, e ApplicationEvent) error方法代替
	// 同步通知事件,不同于PublishEvent，NotifyEvent在Processor Close之后仍然能向Listener发送事件。
	NotifyEvent(e ApplicationEvent) error

	// 启动处理器，如有初始化操作必须定义在该方法
	Start() error

	// 停止处理，与Start方法对应，如有针对Start初始化的清理操作必须定义在该方法
	Close() error
}
