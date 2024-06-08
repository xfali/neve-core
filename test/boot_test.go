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

package test

import (
	"github.com/xfali/neve-core/bean"
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
	err = boot.RegisterBeanByName("x", &aImpl{v: "x value"})
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

	err = boot.RegisterBeanByName("d", func(a *aImpl) *bImpl {
		return &bImpl{custom: a.v}
	})
	if err != nil {
		xlog.Fatal(err)
	}

	err = boot.RegisterBeanByName("struct1", func() a {
		return bImpl{Payload: "this is struct 1"}
	})
	if err != nil {
		xlog.Fatal(err)
	}

	err = boot.RegisterBeanByName("struct2", func() a {
		return bImpl{Payload: "this is struct 2"}
	})
	if err != nil {
		xlog.Fatal(err)
	}

	err = boot.RegisterBeanByName("e", testTmp{}.t)
	if err != nil {
		xlog.Fatal(err)
	}

	err = boot.RegisterBean(bean.NewCustomBeanFactory(func(a *aImpl) *dImpl {
		return &dImpl{custom: a.v}
	}, "DoInit", "DoDestroy"))
	if err != nil {
		xlog.Fatal(err)
	}

	// Singleton
	err = boot.RegisterBeanByName("xx", bean.NewCustomBeanFactoryWithName(bean.SingletonFactory(func(a a) *dImpl {
		return &dImpl{custom: a.Get()}
	}), []string{"x"}, "", ""))
	if err != nil {
		xlog.Fatal(err)
	}

	// Singleton
	err = boot.RegisterBeanByName("xx2", bean.NewCustomBeanFactoryWithName(func(a a) *dImpl {
		return &dImpl{custom: a.Get()}
	}, []string{",omiterror"}, "", ""))
	if err != nil {
		xlog.Fatal(err)
	}

	err = boot.RegisterBean(bean.NewCustomBeanFactoryWithOpts(func(a a) *eImpl {
		return &eImpl{custom: a.Get()}
	}, bean.CustomBeanFactoryOpts.Names([]string{",omiterror"}),
		bean.CustomBeanFactoryOpts.PreAfterSet("PostConstruct"),
		bean.CustomBeanFactoryOpts.PostDestroy("PreDestroy")))
	if err != nil {
		xlog.Fatal(err)
	}

	err = boot.RegisterBean([]string{"hello", "world"})
	if err != nil {
		xlog.Fatal(err)
	}

	err = boot.RegisterBean(&injectBean{})
	if err != nil {
		xlog.Fatal(err)
	}
}

func TestBoot(t *testing.T) {
	go func() {
		time.Sleep(2 * time.Second)
		boot.Stop()
		boot.Stop()
	}()
	err := boot.Run()
	t.Log(err)
}
