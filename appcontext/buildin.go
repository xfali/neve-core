/*
 * Copyright (C) 2024, Xiongfa Li.
 * All rights reserved.
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

type BaseApplicationEvent struct {
	__timestamp time.Time
	__ctx       context.Context
}

func NewBaseApplicationEvent() *BaseApplicationEvent {
	ret := &BaseApplicationEvent{
		__timestamp: time.Now(),
		__ctx:       context.Background(),
	}
	return ret
}

func (e *BaseApplicationEvent) ResetOccurredTime() {
	e.__timestamp = time.Now()
}

func (e *BaseApplicationEvent) OccurredTime() time.Time {
	return e.__timestamp
}

func (e *BaseApplicationEvent) SetEventContext(ctx context.Context) {
	e.__ctx = ctx
}

func (e *BaseApplicationEvent) GetEventContext() context.Context {
	return e.__ctx
}

type ApplicationContextEvent struct {
	BaseApplicationEvent
	appCtx ApplicationContext
}

func (e *ApplicationContextEvent) GetAppContext() ApplicationContext {
	return e.appCtx
}

// 服务启动后触发，Bean已经初始化完成，可以执行任意的业务逻辑
type ContextStartedEvent struct {
	ApplicationContextEvent
}

func NewContextStartedEvent(appCtx ApplicationContext) *ContextStartedEvent {
	ret := &ContextStartedEvent{}
	ret.ResetOccurredTime()
	ret.appCtx = appCtx
	return ret
}

// 服务停止后触发，应尽快做清理工作
type ContextStoppedEvent struct {
	ApplicationContextEvent
}

func NewContextStoppedEvent(appCtx ApplicationContext) *ContextStoppedEvent {
	ret := &ContextStoppedEvent{}
	ret.ResetOccurredTime()
	ret.appCtx = appCtx
	return ret
}

// 已到达ApplicationContext生命周期末端，应用即将退出
type ContextClosedEvent struct {
	ApplicationContextEvent
}

func NewContextClosedEvent(appCtx ApplicationContext) *ContextClosedEvent {
	ret := &ContextClosedEvent{}
	ret.ResetOccurredTime()
	ret.appCtx = appCtx
	return ret
}

// GetEventContext 从事件中提取事件context
// 参数 e: 事件
// 参数 defaultCtx: 默认context，如果e事件实现了EventContextHolder则返回事件的context，否则返回defaultCtx
// 返回 事件context或默认context
func GetEventContext(e ApplicationEvent, defaultCtx context.Context) context.Context {
	if ev, ok := e.(EventContextHolder); ok {
		return ev.GetEventContext()
	}
	return defaultCtx
}
