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

package bean

type Container interface {
	// 注册对象
	// opts添加bean注册的配置，详情查看RegisterOpt
	Register(o interface{}, opts ...RegisterOpt) error

	// 根据名称注册对象
	// opts添加bean注册的配置，详情查看RegisterOpt
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
