// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"github.com/xfali/fig"
	"github.com/xfali/neve-core"
	"github.com/xfali/neve-core/appcontext"
	"github.com/xfali/neve-core/processor"
	"github.com/xfali/neve-utils/neverror"
	"github.com/xfali/xlog"
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
}

func (f *funcBean) RegisterFunction(registry appcontext.InjectFunctionRegistry) error {
	err := registry.RegisterInjectFunction(f.inject)
	if err != nil {
		return err
	}
	err = registry.RegisterInjectFunctionWithNames([]string{"", "a", ""}, f.inject2)
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

func (f *funcBean) inject2(as []a, a a, b *bImpl) {
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
	if a.Get() != "x" {
		xlog.Fatalln("expect: x but get: ", a.Get())
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
