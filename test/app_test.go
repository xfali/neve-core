// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"errors"
	"github.com/xfali/fig"
	"github.com/xfali/neve-core"
	"github.com/xfali/neve-core/container"
	"github.com/xfali/neve-core/processor"
	"github.com/xfali/neve-utils/neverror"
	"github.com/xfali/xlog"
	"testing"
)

type a interface {
	Get() string
}

type aImpl struct {
	v string
}

func (a *aImpl) Get() string {
	return a.v
}

type bImpl struct {
	V string `fig:"userdata.value"`
}

func (b *bImpl) Get() string {
	return b.V
}

func (b *bImpl) AfterSet() error {
	xlog.Infoln("bImpl set")
	return nil
}

type injectBean struct {
	A  a      `inject:""`
	B  a      `inject:"b"`
	BS *bImpl `inject:"b"`
	Bf a      `inject:"c"`
}

func TestApp(t *testing.T) {
	app := neve.NewFileConfigApplication("assets/application-test.yaml")
	err := app.RegisterBean(processor.NewValueProcessor())
	if err != nil {
		t.Fatal(err)
	}
	err = app.RegisterBean(&testProcessor{})
	if err != nil {
		t.Fatal(err)
	}
	err = app.RegisterBean(&aImpl{v: "0"})
	if err != nil {
		t.Fatal(err)
	}
	err = app.RegisterBeanByName("b", &bImpl{})
	if err != nil {
		t.Fatal(err)
	}
	// 注意，此处如果使用RegisterBean会使用a的类型名称注册构造方法
	err = app.RegisterBeanByName("c", func() a {
		return &bImpl{V: "hello world"}
	})
	if err != nil {
		t.Fatal(err)
	}

	err = app.RegisterBean(&injectBean{})
	if err != nil {
		t.Fatal(err)
	}

	app.Run()
}

type testProcessor struct {
	injectBean *injectBean
}

func (p *testProcessor) Init(conf fig.Properties, container container.Container) error {
	return nil
}

func (p *testProcessor) Classify(o interface{}) (bool, error) {
	if x, ok := o.(*bImpl); ok {
		xlog.Infoln("bImpl value is: ", x.V)
	}
	if x, ok := o.(*injectBean); ok {
		p.injectBean = x
	}
	return true, nil
}

func (p *testProcessor) Process() error {
	v := p.injectBean
	if v.A.Get() != "0" {
		xlog.Fatalln("expect: 0 but get: ", v.A.Get())
	}
	if v.B.Get() != "this is a test" {
		xlog.Fatalln("expect: 'this is a test' but get: ", v.B.Get())
	}
	if v.BS.Get() != "this is a test" {
		xlog.Fatalln("expect: 'this is a test' but get: ", v.BS.Get())
	}
	if v.Bf.Get() != "hello world" {
		xlog.Fatalln("expect: 'hello world' but get: ", v.BS.Get())
	}
	xlog.Infoln("all pass, exit")
	//os.Exit(0)
	return nil
}

func (p *testProcessor) Destroy() error {
	xlog.Infoln("testProcessor destroyed")
	return nil
}

func TestPanic(t *testing.T) {
	err := func() (err error) {
		defer neverror.HandleError(&err)
		neverror.PanicError(errors.New("test"))
		return
	}()
	t.Log(err)
}
