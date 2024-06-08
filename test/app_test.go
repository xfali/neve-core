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
	"errors"
	"github.com/xfali/fig"
	"github.com/xfali/neve-core"
	"github.com/xfali/neve-core/appcontext"
	"github.com/xfali/neve-core/bean"
	"github.com/xfali/neve-core/injector"
	"github.com/xfali/neve-core/processor"
	"github.com/xfali/neve-utils/neverror"
	"github.com/xfali/xlog"
	"os"
	"testing"
	"time"
)

type a interface {
	Get() string
}

type aImpl struct {
	v string
}

type testTmp struct {
}

func (t testTmp) t(a *aImpl) *bImpl {
	return &bImpl{custom: a.v}
}

func (a *aImpl) Get() string {
	return a.v
}

type bImpl struct {
	V       string `fig:"userdata.value"`
	custom  string
	Payload interface{}
}

func (b bImpl) Get() string {
	return b.V
}

func (b *bImpl) BeanAfterSet() error {
	xlog.Infof("bImpl set, V: %s custom: %s payload: %v\n", b.V, b.custom, b.Payload)
	if b.V != "this is a test" {
		xlog.Fatalln("b.V not inject")
	}
	return nil
}

func (b *bImpl) BeanDestroy() error {
	xlog.Infoln("bImpl destroyed")
	return nil
}

type c struct {
	V string `value:"userdata.value"`
	I int    `value:"userdata.test"`
}

type dImpl struct {
	V      string `fig:"userdata.value"`
	custom string
}

type eImpl struct {
	V      string `fig:"userdata.value"`
	custom string
}

func (b *dImpl) BeanAfterSet() error {
	xlog.Infof("dImpl BeanAfterSet set, V: %v %p", b.V, b)
	if b.V != "this is a test" {
		xlog.Fatalln("b.V not inject")
	}
	return nil
}

func (b *dImpl) DoInit() error {
	xlog.Infof("dImpl DoInit set, V: %v %p\n", b.V, b)
	if b.V != "this is a test" {
		xlog.Fatalln("b.V not inject")
	}
	return nil
}

func (b *dImpl) DoDestroy() error {
	xlog.Infoln("dImpl DoInit")
	return nil
}

func (e *eImpl) PostConstruct() {
	e.custom = "PostConstruct"
}

func (e *eImpl) BeanAfterSet() error {
	if e.custom != "PostConstruct" {
		xlog.Fatalln(`e.custom != "PostConstruct"`)
	}
	return nil
}

func (e *eImpl) PreDestroy() {
	e.V = "PreDestroy"
}

type order1 struct {
	t *testing.T
	s string
}

func (o *order1) BeanAfterSet() error {
	o.t.Log("order1 set")
	o.s = "order1"
	return nil
}

type order2 struct {
	t *testing.T

	s  string
	O1 *order1 `inject:""`
}

func (o *order2) BeanAfterSet() error {
	o.t.Log("order2 set")
	if o.O1 == nil {
		o.t.Fatalf("cannot nil!")
	}
	o.s = "order2"
	if o.O1.s == "" {
		o.t.Fatalf("cannot empty!")
	}
	o.t.Log(o.O1.s)
	return nil
}

type order3 struct {
	t  *testing.T
	O1 *order1 `inject:""`
	O2 *order2 `inject:""`
}

func (o *order3) BeanAfterSet() error {
	o.t.Log("order3 set")
	if o.O1 == nil {
		o.t.Fatalf("O1 cannot nil!")
	}
	if o.O1.s == "" {
		o.t.Fatalf("O1 cannot empty!")
	}
	o.t.Log(o.O1.s)
	if o.O2 == nil {
		o.t.Fatalf("O2 cannot nil!")
	}
	if o.O2.s == "" {
		o.t.Fatalf("O2 cannot empty!")
	}
	o.t.Log(o.O2.s)
	return nil
}

type testBean interface {
	validate()
}

type injectBean struct {
	A             a        `inject:""`
	B             a        `inject:"b"`
	BS            *bImpl   `inject:"b,omiterror"`
	Bf            a        `inject:"c"`
	Afunc         *bImpl   `inject:"d"`
	Bfunc         *bImpl   `inject:"e"`
	CustomerBean  *dImpl   `inject:""`
	CBwithName    *dImpl   `inject:"xx"`
	CBwithName2   *dImpl   `inject:"xx"`
	CBwithName3   *dImpl   `inject:"xx2"`
	PostConstruct *eImpl   `inject:""`
	Slice         []a      `inject:""`
	Strings       []string `inject:""`
	Struct1       a        `inject:"struct1"`
	Struct2       a        `inject:"struct2"`
}

func (v *injectBean) validate() {
	if v.A.Get() != "0" {
		xlog.Fatalln("expect: 0 but get: ", v.A.Get())
	}
	if v.B.Get() != "this is a test" {
		xlog.Fatalln("expect: 'this is a test' but get: ", v.B.Get())
	}
	if v.BS.Get() != "this is a test" {
		xlog.Fatalln("expect: 'this is a test' but get: ", v.BS.Get())
	}
	if v.Bf.Get() != "this is a test" {
		xlog.Fatalln("expect: 'this is a test' but get: ", v.BS.Get())
	}
	//if v.Bf.Get() != "hello world" {
	//	xlog.Fatalln("expect: 'hello world' but get: ", v.BS.Get())
	//}
	if v.Afunc.custom != "0" {
		xlog.Fatalln("expect: '0' but get: ", v.Afunc.custom)
	} else {
		xlog.Infoln("Afunc: ", v.Afunc.custom)
	}

	if v.Bfunc.custom != "0" || v.Bfunc.V != "this is a test" {
		xlog.Fatalln("expect: '0' but get: ", v.Bfunc.custom)
	} else {
		xlog.Infoln("Bfunc: ", v.Bfunc.custom)
	}

	if v.CustomerBean.custom != "0" || v.CustomerBean.V != "this is a test" {
		xlog.Fatalln("expect: '0' but get: ", v.CustomerBean.custom)
	} else {
		xlog.Infoln("CustomerBean: ", v.CustomerBean.custom)
	}

	if v.CBwithName.custom != "x value" || v.CBwithName.V != "this is a test" {
		xlog.Fatalln("expect: 'x value' but get: ", v.CBwithName.custom)
	} else {
		xlog.Infoln("CBwithName: ", v.CBwithName.custom)
	}

	if v.CBwithName != v.CBwithName2 {
		xlog.Fatalf("expect: CBwithName %p equal CBwithName2 %p but not", v.CBwithName, v.CBwithName2)
	} else {
		xlog.Infof("CBwithName: %p  CBwithName2 %p", v.CBwithName, v.CBwithName2)
	}
	if v.CBwithName3.custom != v.A.Get() {
		xlog.Fatalf("expect: CBwithName3 %s equal A %s but not", v.CBwithName3.custom, v.A.Get())
	} else {
		xlog.Infof("CBwithName3: %s  A %s", v.CBwithName3.custom, v.A.Get())
	}

	if v.PostConstruct.custom != "PostConstruct" {
		xlog.Fatalf("expect: PostConstruct %s equal PostConstruct %s but not", v.PostConstruct.V, "PostConstruct")
	} else {
		xlog.Infof("PostConstruct: %s  A %s", v.PostConstruct.V, "PostConstruct")
	}

	//if v.PostConstruct.V != "PreDestroy" {
	//	xlog.Fatalf("expect: PreDestroy %s equal PreDestroy %s but not", v.PostConstruct.V, "PreDestroy")
	//} else {
	//	xlog.Infof("PreDestroy: %s  A %s", v.PostConstruct.V, "PreDestroy")
	//}

	if len(v.Slice) == 0 {
		xlog.Fatalln("expect larger than 0, but get ", len(v.Slice))
	} else {
		for _, a := range v.Slice {
			xlog.Infoln(a.Get())
		}
	}

	if len(v.Strings) == 0 {
		xlog.Fatalln("expect larger than 0, but get ", len(v.Strings))
	} else {
		for _, a := range v.Strings {
			xlog.Infoln(a)
		}
	}

	xlog.Infoln(v.Struct1, v.Struct2)
}

type injectBeanB struct {
	A  a      `Autowired:""`
	B  a      `Autowired:"b"`
	BS *bImpl `Autowired:"b"`
	Bf a      `Autowired:"c"`
}

func (v *injectBeanB) validate() {
	if v.A.Get() != "0" {
		xlog.Fatalln("expect: 0 but get: ", v.A.Get())
	}
	if v.B.Get() != "this is a test" {
		xlog.Fatalln("expect: 'this is a test' but get: ", v.B.Get())
	}
	if v.BS.Get() != "this is a test" {
		xlog.Fatalln("expect: 'this is a test' but get: ", v.BS.Get())
	}
	if v.Bf.Get() != "this is a test" {
		xlog.Fatalln("expect: 'this is a test' but get: ", v.BS.Get())
	}
}

func TestApp(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		app := neve.NewFileConfigApplication("assets/application-test.yaml")
		o := &injectBean{}
		testApp(app, t, o)
		if o.A == nil || o.B == nil || o.Bf == nil || o.BS == nil {
			t.Fatal("not match")
		}
	})

	t.Run("changeTag", func(t *testing.T) {
		app := neve.NewFileConfigApplication("assets/application-test.yaml",
			neve.OptSetInjectTagName("Autowired"))
		o := &injectBeanB{}
		testApp(app, t, o)
		if o.A == nil || o.B == nil || o.Bf == nil || o.BS == nil {
			t.Fatal("not match")
		}
	})

	t.Run("change default tag", func(t *testing.T) {
		injector.InjectTagName = "Autowired"
		app := neve.NewFileConfigApplication("assets/application-test.yaml")
		o := &injectBeanB{}
		testApp(app, t, o)
		if o.A == nil || o.B == nil || o.Bf == nil || o.BS == nil {
			t.Fatal("not match")
		}
	})
}

func testApp(app neve.Application, t *testing.T, o interface{}) {
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
	err = app.RegisterBeanByName("x", &aImpl{v: "x value"})
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

	err = app.RegisterBeanByName("d", func(a *aImpl) *bImpl {
		return &bImpl{custom: a.v}
	})
	if err != nil {
		t.Fatal(err)
	}

	err = app.RegisterBeanByName("struct1", func() a {
		return bImpl{Payload: "this is struct 1"}
	})
	if err != nil {
		t.Fatal(err)
	}

	err = app.RegisterBeanByName("struct2", func() a {
		return bImpl{Payload: "this is struct 2"}
	})
	if err != nil {
		t.Fatal(err)
	}

	err = app.RegisterBeanByName("e", testTmp{}.t)
	if err != nil {
		t.Fatal(err)
	}

	err = app.RegisterBean(bean.NewCustomBeanFactory(func(a *aImpl) *dImpl {
		return &dImpl{custom: a.v}
	}, "DoInit", "DoDestroy"))
	if err != nil {
		t.Fatal(err)
	}

	// Singleton
	err = app.RegisterBeanByName("xx", bean.NewCustomBeanFactoryWithName(bean.SingletonFactory(func(a a) *dImpl {
		return &dImpl{custom: a.Get()}
	}), []string{"x"}, "", ""))
	if err != nil {
		t.Fatal(err)
	}

	// Singleton
	err = app.RegisterBeanByName("xx2", bean.NewCustomBeanFactoryWithName(func(a a) *dImpl {
		return &dImpl{custom: a.Get()}
	}, []string{",omiterror"}, "", ""))
	if err != nil {
		t.Fatal(err)
	}

	err = app.RegisterBean(bean.NewCustomBeanFactoryWithOpts(func(a a) *eImpl {
		return &eImpl{custom: a.Get()}
	}, bean.CustomBeanFactoryOpts.Names([]string{",omiterror"}),
		bean.CustomBeanFactoryOpts.PreAfterSet("PostConstruct"),
		bean.CustomBeanFactoryOpts.PostDestroy("PreDestroy")))
	if err != nil {
		t.Fatal(err)
	}

	err = app.RegisterBean([]string{"hello", "world"})
	if err != nil {
		t.Fatal(err)
	}

	err = app.RegisterBean(o)
	if err != nil {
		t.Fatal(err)
	}

	go app.Run()
	time.Sleep(2 * time.Minute)
}

func TestAppCircleDependency(t *testing.T) {
	app := neve.NewFileConfigApplication("assets/application-test.yaml")
	type testT struct {
		A *bImpl `inject:""`
	}
	err := app.RegisterBean(processor.NewValueProcessor())
	if err != nil {
		t.Fatal(err)
	}
	err = app.RegisterBean(&testProcessor{})
	if err != nil {
		t.Fatal(err)
	}
	o := &testT{}
	err = app.RegisterBean(o)
	if err != nil {
		t.Fatal(err)
	}
	err = app.RegisterBeanByName("", func(a *testT, b *bImpl) *bImpl {
		return a.A
	})
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		defer func() {
			o := recover()
			if o == nil {
				t.Fatal("Must be panic Circular dependency")
			} else {
				t.Log(o)
			}
		}()
		app.Run()
	}()
	time.Sleep(2 * time.Second)
	if o.A != nil {
		t.Fatal("expect nil but get ", o.A)
	}
}

func TestValue(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		app := neve.NewFileConfigApplication("assets/application-test.yaml")
		o := &injectBean{}
		testvalue(app, t, o)
	})

	t.Run("change tag name", func(t *testing.T) {
		app := neve.NewFileConfigApplication("assets/application-test.yaml",
			neve.OptSetInjectTagName("Autowired"))
		o := &injectBeanB{}
		testvalue(app, t, o)
	})
}

func testvalue(app neve.Application, t *testing.T, o interface{}) {
	err := app.RegisterBean(processor.NewValueProcessor(processor.OptSetValueTag("valuePx", "value")))
	if err != nil {
		t.Fatal(err)
	}

	v := &c{}
	err = app.RegisterBean(v)
	if err != nil {
		t.Fatal(err)
	}
	go app.Run()

	time.Sleep(2 * time.Second)

	t.Log(v.V)
	if v.V != "this is a test" {
		t.Fatalf("not match")
	}

	t.Log(v.I)
	if v.I != 100 {
		t.Fatalf("not match")
	}
}

type testProcessor struct {
	injectBean testBean
}

func (p *testProcessor) Init(conf fig.Properties, container bean.Container) error {
	return nil
}

func (p *testProcessor) Classify(o interface{}) (bool, error) {
	if x, ok := o.(*bImpl); ok {
		xlog.Infoln("bImpl value is: ", x.V)
	}
	if x, ok := o.(testBean); ok {
		p.injectBean = x
	}
	return true, nil
}

func (p *testProcessor) Process() error {
	v := p.injectBean
	if v != nil {
		v.validate()
	}
	xlog.Infoln("all pass, exit")
	os.Exit(0)
	return nil
}

func (p *testProcessor) BeanDestroy() error {
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

func TestOrder(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		app := neve.NewFileConfigApplication("assets/application-test.yaml")
		testOrder(app, t)
	})
}

func testOrder(app neve.Application, t *testing.T) {
	err := app.RegisterBean(processor.NewValueProcessor())
	if err != nil {
		t.Fatal(err)
	}
	err = app.RegisterBean(&order2{t: t}, neve.SetOrder(2))
	if err != nil {
		t.Fatal(err)
	}
	err = app.RegisterBeanByName("c", &order3{t: t}, neve.SetOrder(3))
	if err != nil {
		t.Fatal(err)
	}

	err = app.RegisterBean(&order1{t: t}, bean.SetOrder(-1))
	if err != nil {
		t.Fatal(err)
	}

	go app.Run()
	time.Sleep(2 * time.Second)
}

type listener struct {
	t *testing.T
}

type customerEvent struct {
	appcontext.BaseApplicationEvent

	payload string
}

func newCustomerEvent(payload string) *customerEvent {
	ret := &customerEvent{
		BaseApplicationEvent: *appcontext.NewBaseApplicationEvent(),
	}
	ret.payload = payload
	return ret
}

func (l *listener) Event(event appcontext.ApplicationEvent) {
	//	l.t.Log(reflection.GetObjectName(event))
	if e, ok := event.(*customerEvent); ok {
		if e.payload != "hello world" {
			l.t.Fatal("not match")
		}
		l.t.Log("func event:", e.payload)
	}
}

func (l *listener) OnApplicationEvent(event appcontext.ApplicationEvent) {
	//	l.t.Log(reflection.GetObjectName(event))
	if e, ok := event.(*customerEvent); ok {
		if e.payload != "hello world" {
			l.t.Fatal("not match")
		}
		l.t.Log(e.payload)
	}
}

func (l *listener) EventStarted(event *appcontext.ContextStartedEvent) {
	l.t.Log("ContextStartedEvent:", event.OccurredTime())
	event.GetAppContext().PublishEvent(newCustomerEvent("hello world"))
	event.GetAppContext().PublishEvent(appcontext.NewPayloadApplicationEvent("hello world"))
	event.GetAppContext().PublishEvent(appcontext.NewPayloadApplicationEvent(&aImpl{v: "hello world2"}))
}

func (l *listener) EventStopped(event *appcontext.ContextStoppedEvent) {
	l.t.Log("ContextStoppedEvent:", event.OccurredTime())
}

func (l *listener) payload(payload string) {
	if payload != "hello world" {
		l.t.Fatal("not match")
	}
	l.t.Log("listener payload", payload)
}

func (l *listener) payloadA(payload a) {
	if payload.Get() != "hello world2" {
		l.t.Fatal("not match")
	}
	l.t.Log("listener payloadA", payload.Get())
}

type listener2 struct {
	appcontext.PayloadEventListener
	t *testing.T
}

func (l *listener2) payloadA(payload a) {
	if payload.Get() != "hello world2" {
		l.t.Fatal("not match")
	}
	l.t.Log("listener2 payloadA", payload.Get())
}

type listener3 struct {
	t *testing.T
}

func (l *listener3) RegisterConsumer(register appcontext.ApplicationEventConsumerRegistry) error {
	return register.RegisterApplicationEventConsumer(l.handlerEvent)
}

func (l *listener3) handlerEvent(event *customerEvent) {
	if event.payload != "hello world" {
		l.t.Fatal("not match")
	}
	l.t.Log("listener3", event.payload)
}

type publisher struct {
	t *testing.T

	Publisher appcontext.ApplicationEventPublisher `inject:""`
}

func (l *publisher) BeanAfterSet() error {
	if l.Publisher == nil {
		l.t.Fatal("Publisher is nil")
	}
	l.t.Log("publisher publish  newCustomerEvent hello world ===================")
	return l.Publisher.PublishEvent(newCustomerEvent("hello world"))
}

func TestListener(t *testing.T) {
	t.Run("default enable", func(t *testing.T) {
		app := neve.NewFileConfigApplication("assets/application-test.yaml")
		testListener(app, t)
		time.Sleep(2 * time.Second)
	})

	t.Run("default disable", func(t *testing.T) {
		defer func() {
			o := recover()
			t.Log(o)
		}()
		app := neve.NewFileConfigApplication("assets/application-test.yaml",
			neve.OptSetApplicationContext(appcontext.NewDefaultApplicationContext(appcontext.OptDisableEvent())))
		testListener(app, t)
		time.Sleep(2 * time.Second)
	})
}

func testListener(app neve.Application, t *testing.T) {
	f := &listener{t: t}
	app.AddListeners(
		f,
		f.Event,
		f.EventStarted,
		f.EventStopped,
	)
	err := app.RegisterBean(appcontext.NewPayloadEventListener(f.payload))
	if err != nil {
		t.Fatal(err)
	}
	err = app.RegisterBeanByName("A", appcontext.NewPayloadEventListener(f.payloadA))
	if err != nil {
		t.Fatal(err)
	}

	l2 := &listener2{}
	l2.t = t
	l2.RefreshPayloadHandler(l2.payloadA)
	err = app.RegisterBean(l2)
	if err != nil {
		t.Fatal(err)
	}

	err = app.RegisterBeanByName("C", appcontext.NewPayloadEventListener(f.payload, f.payloadA, l2.payloadA))
	if err != nil {
		t.Fatal(err)
	}

	l3 := &listener3{}
	l3.t = t
	err = app.RegisterBean(l3)
	if err != nil {
		t.Fatal(err)
	}

	pub := &publisher{}
	pub.t = t
	err = app.RegisterBean(pub)
	if err != nil {
		t.Fatal(err)
	}

	go app.Run()
}

type aware struct {
	t *testing.T
}

func (a *aware) SetApplicationContext(ctx appcontext.ApplicationContext) {
	if ctx == nil {
		a.t.Fatal("must not be nil")
	}
	a.t.Log(ctx.GetApplicationName())
}

func TestAware(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		app := neve.NewFileConfigApplication("assets/application-test.yaml")
		testAware(app, t)
		time.Sleep(2 * time.Second)
	})
}

func testAware(app neve.Application, t *testing.T) {
	err := app.RegisterBean(&aware{t: t})
	if err != nil {
		t.Fatal(err)
	}

	go app.Run()
}
