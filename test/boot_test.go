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

package test

import (
	"github.com/xfali/neve-core/boot"
	"github.com/xfali/neve-core/processor"
	"github.com/xfali/xlog"
	"testing"
	"time"
)

func init() {
	testing.Init()
	boot.ConfigPath = "assets/application-test.yaml"
	err := boot.RegisterBean(processor.NewValueProcessor())
	if err != nil {
		xlog.Fatal(err)
	}
	err = boot.RegisterBean(&testProcessor{})
	if err != nil {
		xlog.Fatal(err)
	}
	err = boot.RegisterBean(&aImpl{v: "0"})
	if err != nil {
		xlog.Fatal(err)
	}
	err = boot.RegisterBeanByName("b", &bImpl{})
	if err != nil {
		xlog.Fatal(err)
	}
	// 注意，此处如果使用RegisterBean会使用a的类型名称注册构造方法
	err = boot.RegisterBeanByName("c", func() a {
		return &bImpl{V: "hello world"}
	})
	if err != nil {
		xlog.Fatal(err)
	}

	err = boot.RegisterBean(&injectBean{})
	if err != nil {
		xlog.Fatal(err)
	}
}

func TestBoot(t *testing.T) {
	go boot.Run()
	time.Sleep(2 * time.Second)
}
