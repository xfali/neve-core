// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package inject

import (
	"github.com/xfali/neve/neve-core/container"
	"github.com/xfali/neve/neve-core/injector"
	"io"
	"testing"
)

// 本测试中为了测试未注入的情况，设置了异常的数据，注入过程中会打印Error未正常现象，以最终pass数量为准。

type a interface {
	Get() int
}

type aImpl struct {
}

func (a *aImpl) Get() int {
	return 1
}

type bImpl struct {
	i int
}

func (a *bImpl) Get() int {
	if a.i != 0 {
		return a.i
	}
	return 2
}

type dest struct {
	A a `inject:""`
	B a `inject:"b"`
	// Would not inject
	C io.Writer `inject:""`
}

func TestInjectInterface(t *testing.T) {
	t.Run("inject once", func(t *testing.T) {
		c := container.New()
		c.Register(&aImpl{})
		c.RegisterByName("b", &bImpl{})
		i := injector.New()

		d := dest{}
		err := i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		if d.A == nil || d.A.Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.B == nil || d.B.Get() != 2 {
			t.Fatal("inject B failed")
		}
	})

	t.Run("inject twice", func(t *testing.T) {
		c := container.New()
		c.Register(&aImpl{})
		c.RegisterByName("b", &bImpl{})
		i := injector.New()

		d := dest{}
		err := i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		if d.A == nil || d.A.Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.B == nil || d.B.Get() != 2 {
			t.Fatal("inject B failed")
		}

		err = i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		if d.A == nil || d.A.Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.B == nil || d.B.Get() != 2 {
			t.Fatal("inject B failed")
		}
	})

	t.Run("inject twice with modify", func(t *testing.T) {
		c := container.New()
		c.Register(&aImpl{})
		b := &bImpl{}
		c.RegisterByName("b", b)
		i := injector.New()

		d := dest{}
		err := i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		//modify here
		b.i = 3
		if d.A == nil || d.A.Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.B == nil || d.B.Get() != 3 {
			t.Fatal("inject B failed")
		}

		err = i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		b.i = 2
		if d.A == nil || d.A.Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.B == nil || d.B.Get() != 2 {
			t.Fatal("inject B failed")
		}
	})
}

type dest2 struct {
	A  aImpl  `inject:""`
	B  *bImpl `inject:"b"`
	B2 bImpl  `inject:"b"`
	// Would not inject
	C dest `inject:""`
}

func TestInjectStruct(t *testing.T) {
	t.Run("inject once", func(t *testing.T) {
		c := container.New()
		c.Register(&aImpl{})
		c.RegisterByName("b", &bImpl{})
		i := injector.New()

		d := dest2{}
		err := i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		if d.A.Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.B.Get() != 2 {
			t.Fatal("inject B failed")
		}
	})

	t.Run("inject twice", func(t *testing.T) {
		c := container.New()
		c.Register(&aImpl{})
		c.RegisterByName("b", &bImpl{})
		i := injector.New()

		d := dest2{}
		err := i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		if d.A.Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.B.Get() != 2 {
			t.Fatal("inject B failed")
		}

		err = i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		if d.A.Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.B.Get() != 2 {
			t.Fatal("inject B failed")
		}
	})

	t.Run("inject twice with modify", func(t *testing.T) {
		c := container.New()

		c.Register(&aImpl{})
		b := &bImpl{}
		c.RegisterByName("b", b)
		i := injector.New()

		d := dest2{}
		err := i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		b.i = 3
		if d.A.Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.B.Get() != 3 {
			t.Fatal("inject B failed")
		}
		if d.B2.Get() != 2 {
			t.Fatal("inject B2 failed")
		}

		err = i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		b.i = 2
		if d.A.Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.B.Get() != 2 {
			t.Fatal("inject B failed")
		}
		//if d.B2.Get() != 3 {
		//	t.Fatal("inject B2 failed")
		//}
		// struct只允许注入指针类型
		if d.B2.Get() != 2 {
			t.Fatal("inject B2 failed")
		}
	})
}

func TestInjectFunc(t *testing.T) {
	f1 := func() a {
		return &aImpl{}
	}
	f2 := func() *bImpl {
		return &bImpl{}
	}
	t.Run("inject once", func(t *testing.T) {
		c := container.New()
		c.Register(f1)
		c.RegisterByName("b", f2)
		i := injector.New()

		d := dest{}
		err := i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		if d.A.Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.B.Get() != 2 {
			t.Fatal("inject B failed")
		}
	})

	t.Run("inject twice", func(t *testing.T) {
		f1 := func() a {
			return &aImpl{}
		}
		f2 := func() *bImpl {
			return &bImpl{}
		}
		c := container.New()
		c.Register(f1)
		c.RegisterByName("b", f2)
		i := injector.New()

		d := dest2{}
		err := i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		if d.A.Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.B.Get() != 2 {
			t.Fatal("inject B failed")
		}

		err = i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		if d.A.Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.B.Get() != 2 {
			t.Fatal("inject B failed")
		}
	})

	t.Run("inject interface twice with modify", func(t *testing.T) {
		x := 0
		x1 := &x
		f1 := func() a {
			return &aImpl{}
		}
		f2 := func() *bImpl {
			return &bImpl{i: *x1}
		}

		c := container.New()

		c.Register(f1)
		c.RegisterByName("b", f2)
		i := injector.New()

		d := dest{}
		//必须在注入前修改
		*x1 = 3
		err := i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		if d.A.Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.B.Get() != 3 {
			t.Fatal("inject B failed")
		}

		*x1 = 2
		err = i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		if d.A.Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.B.Get() != 2 {
			t.Fatal("inject B failed")
		}
	})

	t.Run("inject strcut twice with modify", func(t *testing.T) {
		x := 0
		x1 := &x
		f1 := func() a {
			return &aImpl{}
		}
		f2 := func() *bImpl {
			return &bImpl{i: *x1}
		}

		c := container.New()

		c.Register(f1)
		c.RegisterByName("b", f2)
		i := injector.New()

		d := dest2{}
		//必须在注入前修改
		*x1 = 3
		err := i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		if d.A.Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.B.Get() != 3 {
			t.Fatal("inject B failed")
		}
		if d.B2.Get() != 2 {
			t.Fatal("inject B2 failed")
		}

		*x1 = 2
		err = i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		if d.A.Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.B.Get() != 2 {
			t.Fatal("inject B failed")
		}
		//if d.B2.Get() != 3 {
		//	t.Fatal("inject B2 failed")
		//}
		// struct只允许注入指针类型
		if d.B2.Get() != 2 {
			t.Fatal("inject B2 failed")
		}
	})
}
