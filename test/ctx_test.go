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
