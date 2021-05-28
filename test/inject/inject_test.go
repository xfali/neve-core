// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package inject

import (
	"github.com/xfali/neve-core/bean"
	"github.com/xfali/neve-core/injector"
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
		c := bean.NewContainer()
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
		c := bean.NewContainer()
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
		c := bean.NewContainer()
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

type slice struct {
	A []a      `inject:""`
	B []a      `inject:""`
	C []a      `inject:"test"`
	D []*bImpl `inject:"test"`
}

func TestInjectStruct(t *testing.T) {
	t.Run("inject once", func(t *testing.T) {
		c := bean.NewContainer()
		c.Register(&aImpl{})
		c.RegisterByName("b", &bImpl{i:3})
		i := injector.New()

		d := dest2{}
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
	})

	t.Run("inject twice", func(t *testing.T) {
		c := bean.NewContainer()
		c.Register(&aImpl{})
		c.RegisterByName("b",  &bImpl{i:3})
		i := injector.New()

		d := dest2{}
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

		err = i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		if d.A.Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.B.Get() != 3 {
			t.Fatal("inject B failed")
		}
	})

	t.Run("inject twice with modify", func(t *testing.T) {
		c := bean.NewContainer()

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

func TestInjectSlice(t *testing.T) {
	t.Run("inject once", func(t *testing.T) {
		c := bean.NewContainer()
		c.Register(&aImpl{})
		c.RegisterByName("b", &bImpl{})
		c.RegisterByName("test", []*bImpl{
			&bImpl{},
		})
		i := injector.New()

		d := slice{}
		err := i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		if d.A[0].Get() != 1 {
			t.Fatal("inject A failed")
		}
		if d.A[1].Get() != 2 {
			t.Fatal("inject A failed")
		}
		if d.B[0].Get() != 1 {
			t.Fatal("inject B failed")
		}
		if d.B[1].Get() != 2 {
			t.Fatal("inject B failed")
		}
		if d.C[0].Get() != 2 {
			t.Fatal("inject C failed")
		}
		if d.D[0].Get() != 2 {
			t.Fatal("inject D failed")
		}
	})

	t.Run("inject twice", func(t *testing.T) {
		c := bean.NewContainer()
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
		c := bean.NewContainer()

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

type testMap struct {
	A map[string]a      `inject:""`
	B map[string]a      `inject:""`
	C map[string]a      `inject:"test"`
	D map[string]*bImpl `inject:"test"`
}

func TestInjectMap(t *testing.T) {
	t.Run("inject once", func(t *testing.T) {
		c := bean.NewContainer()
		c.Register(&aImpl{})
		c.RegisterByName("b", &bImpl{})
		c.RegisterByName("test", map[string]*bImpl{
			"test": &bImpl{},
		})
		i := injector.New()

		d := testMap{
			A: map[string]a{},
			B: map[string]a{},
			C: map[string]a{},
			D: map[string]*bImpl{},
		}
		err := i.Inject(c, &d)
		if err != nil {
			t.Fatal(err)
		}

		for k, v := range d.A {
			t.Logf("key: %s, value: %d", k, v.Get())
			if v.Get() != 1 && v.Get() != 2 {
				t.Fatal("inject A failed")
			}
		}
		for k, v := range d.B {
			t.Logf("key: %s, value: %d", k, v.Get())
			if v.Get() != 1 && v.Get() != 2 {
				t.Fatal("inject B failed")
			}
		}
		for k, v := range d.C {
			t.Logf("key: %s, value: %d", k, v.Get())
			if v.Get() != 2 {
				t.Fatal("inject C failed")
			}
		}
		for k, v := range d.D {
			t.Logf("key: %s, value: %d", k, v.Get())
			if v.Get() != 2 {
				t.Fatal("inject D failed")
			}
		}
	})

	t.Run("inject twice", func(t *testing.T) {
		c := bean.NewContainer()
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
		c := bean.NewContainer()

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
		c := bean.NewContainer()
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
		c := bean.NewContainer()
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

		c := bean.NewContainer()

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

		c := bean.NewContainer()

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

func checkErr(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestInjectComplex(t *testing.T) {
	type injectBean struct {
		A   a      `inject:""`
		B   a      `inject:"b"`
		Bs  *bImpl `inject:"b"`
		Bf  a      `inject:""`
		Bfs *bImpl `inject:"c"`
	}

	dest := &injectBean{}
	c := bean.NewContainer()
	i := injector.New()

	// 此处注册名称为aImpl类型名
	checkErr(c.Register(&aImpl{}), t)
	// 此处注册名称为b类型名
	checkErr(c.RegisterByName("b", &bImpl{i: 2}), t)
	//注意：此处注册的名称为a的类型名
	checkErr(c.Register(func() a {
		return &bImpl{i: 10}
	}), t)
	// 此处注册名称为c
	checkErr(c.RegisterByName("c", func() *bImpl {
		return &bImpl{i: 11}
	}), t)

	var errA a = &bImpl{i: 3}
	// 此处注册名称为bImpl类型名称
	checkErr(c.Register(errA), t)

	checkErr(i.Inject(c, dest), t)

	if dest.A.Get() != 10 {
		t.Fatal("expect 10 but get: ", dest.A.Get())
	}

	if dest.B.Get() != 2 {
		t.Fatal("expect 2 but get: ", dest.B.Get())
	}

	if dest.Bs.Get() != 2 {
		t.Fatal("expect 2 but get: ", dest.Bs.Get())
	}

	if dest.Bf.Get() != 10 {
		t.Fatal("expect 10 but get: ", dest.Bs.Get())
	}

	if dest.Bfs.Get() != 11 {
		t.Fatal("expect 11 but get: ", dest.Bs.Get())
	}
}

type destChange struct {
	A a `Autowired:""`
	B a `Autowired:"b"`
	// Would not inject
	C io.Writer `Autowired:""`
}
func TestInjectTagChange(t *testing.T) {
	t.Run("inject once", func(t *testing.T) {
		c := bean.NewContainer()
		c.Register(&aImpl{})
		c.RegisterByName("b", &bImpl{})
		i := injector.New(injector.OptSetInjectTagName("Autowired"))

		d := destChange{}
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

		if d.C != nil {
			t.Fatal("inject C must failed")
		}
	})
}
