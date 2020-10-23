// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"github.com/xfali/fig"
	"github.com/xfali/neve-core/appcontext"
	"testing"
)

func TestContext(t *testing.T) {
	conf, err := fig.LoadYamlFile("assets/application-test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	ctx := appcontext.NewDefaultApplicationContext(conf)
	ctx.RegisterBean(&bImpl{})
	ctx.Close()
	ctx.Close()
}
