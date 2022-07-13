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
	"github.com/xfali/fig"
	"github.com/xfali/neve-core"
	"github.com/xfali/neve-core/appcontext"
	"github.com/xfali/neve-core/processor"
	"github.com/xfali/neve-utils/neverror"
	"github.com/xfali/xlog"
	"io"
	"testing"
	"time"
)

func TestContext(t *testing.T) {
	conf, err := fig.LoadYamlFile("assets/application-test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	ctx := appcontext.NewDefaultApplicationContext()
	ctx.Init(conf)
	ctx.RegisterBean(&bImpl{})
	ctx.Close()
	ctx.Close()
}

func TestContext2(t *testing.T) {
	ctx := appcontext.NewDefaultApplicationContext()
	app := neve.NewFileConfigApplication("assets/application-test.yaml", neve.OptSetApplicationContext(ctx))
	if app == nil {
		t.Fatal("app is nil")
	}

	err := app.RegisterBean(processor.NewValueProcessor())
	if err != nil {
		t.Fatal(err)
	}
	app.RegisterBean(&bImpl{})
	app.RegisterBean(&injectBean{})
	go func() {
		time.Sleep(time.Second)
		ctx.Close()
		ctx.Close()
	}()
	app.Run()
}

type funcBean struct {
	t *testing.T
	a a
}

func (f *funcBean) BeanAfterSet() error {
	if f.a == nil {
		f.t.Fatal("a is nil")
	}
	f.t.Log("BeanAfterSet: ", f.a.Get())
	return nil
}

func (f *funcBean) RegisterFunction(registry appcontext.InjectFunctionRegistry) error {
	err := registry.RegisterInjectFunction(f.inject)
	if err != nil {
		return err
	}
	err = registry.RegisterInjectFunction(func(as []a, a a, b *bImpl) {
		f.testInject(as, a, b)
	}, "", "a", "")
	if err != nil {
		return err
	}

	err = registry.RegisterInjectFunction(func(w io.Writer) {
	}, ",omiterror")
	if err != nil {
		return err
	}
	return nil
}

func (f *funcBean) inject(as []a, a *aImpl, b *bImpl) {
	if len(as) != 3 {
		f.t.Fatal("as is: ", len(as))
	}
	if as[0].Get() != "0" {
		f.t.Fatal("expect: 0 but get: ", as[0].Get())
	} else {
		f.t.Log("inject func: as[0]: ", as[0].Get())
	}
	if as[1].Get() != "this is a test" {
		f.t.Fatal("expect: 'this is a test' but get: ", as[1].Get())
	} else {
		f.t.Log("inject func: as[1]: ", as[1].Get())
	}

	if a == nil {
		f.t.Fatal("a is nil")
	}
	f.a = a
	if a.Get() != "0" {
		xlog.Fatalln("expect: 0 but get: ", a.Get())
	} else {
		f.t.Log("inject func: a: ", a.Get())
	}

	if b == nil {
		f.t.Fatal("b is nil")
	}
	if b.Get() != "this is a test" {
		f.t.Fatal("expect: 'this is a test' but get: ", b.Get())
	} else {
		f.t.Log("inject func: b: ", b.Get())
	}
}

func (f *funcBean) testInject(as []a, a a, b *bImpl) {
	if len(as) != 3 {
		f.t.Fatal("as is: ", len(as))
	}
	if as[0].Get() != "0" {
		f.t.Fatal("expect: 0 but get: ", as[0].Get())
	} else {
		f.t.Log("testInject inject func: as[0]: ", as[0].Get())
	}
	if as[1].Get() != "this is a test" {
		f.t.Fatal("expect: 'this is a test' but get: ", as[1].Get())
	} else {
		f.t.Log("testInject inject func: as[1]: ", as[1].Get())
	}

	if a == nil {
		f.t.Fatal("a is nil")
	}
	f.a = a
	if a.Get() != "x" {
		xlog.Fatalln("expect: x but get: ", a.Get())
	} else {
		f.t.Log("testInject inject func: a: ", a.Get())
	}

	if b == nil {
		f.t.Fatal("b is nil")
	}
	if b.Get() != "this is a test" {
		f.t.Fatal("expect: 'this is a test' but get: ", b.Get())
	} else {
		f.t.Log("testInject inject func: b: ", b.Get())
	}
}

func TestInjectFunction(t *testing.T) {
	ctx := appcontext.NewDefaultApplicationContext()
	app := neve.NewFileConfigApplication("assets/application-test.yaml", neve.OptSetApplicationContext(ctx))
	if app == nil {
		t.Fatal("app is nil")
	}

	neverror.PanicError(app.RegisterBean(processor.NewValueProcessor()))
	neverror.PanicError(app.RegisterBean(&funcBean{t: t}))
	neverror.PanicError(app.RegisterBean(&aImpl{v: "0"}))
	neverror.PanicError(app.RegisterBean(&bImpl{V: "1"}))
	neverror.PanicError(app.RegisterBeanByName("a", &aImpl{v: "x"}))

	go app.Run()
	time.Sleep(time.Second)
	ctx.Close()
	ctx.Close()
}
