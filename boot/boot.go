/*
 * Copyright (C) 2022, Xiongfa Li.
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

package boot

import (
	"flag"
	"github.com/xfali/neve-core"
	"github.com/xfali/neve-core/bean"
	"sync"
)

var (
	// 默认的配置路径
	ConfigPath = "application.yaml"

	creator func() neve.Application = defaultCreator
	gApp    neve.Application
	once    sync.Once
)

func init() {
}

// 注册到全局Application
// 注册对象
// 支持注册
//  1、interface、struct指针，注册名称使用【类型名称】；
//  2、struct/interface的构造函数 func() TYPE，注册名称使用【返回值的类型名称】。
// opts添加bean注册的配置，详情查看bean.RegisterOpt
func RegisterBean(o interface{}, opts ...bean.RegisterOpt) error {
	return instance().RegisterBean(o, opts...)
}

// 注册到全局Application
// 使用指定名称注册对象
// 支持注册
//  1、interface、struct指针，注册名称使用【类型名称】；
//  2、struct/interface的构造函数 func() TYPE，注册名称使用【返回值的类型名称】。
// opts添加bean注册的配置，详情查看bean.RegisterOpt
func RegisterBeanByName(name string, o interface{}, opts ...bean.RegisterOpt) error {
	return instance().RegisterBeanByName(name, o, opts...)
}

// 自定义启动的Application
// 必须在注册对象和Run之前调用
func Customize(app neve.Application) {
	creator = func() neve.Application {
		return app
	}
}

func defaultCreator() neve.Application {
	flag.StringVar(&ConfigPath, "f", ConfigPath, "Application configuration file path.")
	flag.Parse()
	return neve.NewFileConfigApplication(ConfigPath)
}

func instance() neve.Application {
	once.Do(func() {
		gApp = creator()
	})
	return gApp
}

// 启动全局Application
func Run() error {
	return instance().Run()
}
