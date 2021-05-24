// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

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

func (a *aImpl) Get() string {
	return a.v
}

type bImpl struct {
	V string `fig:"userdata.value"`
}

func (b *bImpl) Get() string {
	return b.V
}

func (b *bImpl) BeanAfterSet() error {
	xlog.Infoln("bImpl set, V: ", b.V)
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
	A  a      `inject:""`
	B  a      `inject:"b"`
	BS *bImpl `inject:"b"`
	Bf a      `inject:"c"`
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
	if v.Bf.Get() != "hello world" {
		xlog.Fatalln("expect: 'hello world' but get: ", v.BS.Get())
	}
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
	if v.Bf.Get() != "hello world" {
		xlog.Fatalln("expect: 'hello world' but get: ", v.BS.Get())
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
	// 注意，此处如果使用RegisterBean会使用a的类型名称注册构造方法
	err = app.RegisterBeanByName("c", func() a {
		return &bImpl{V: "hello world"}
	})
	if err != nil {
		t.Fatal(err)
	}

	err = app.RegisterBean(o)
	if err != nil {
		t.Fatal(err)
	}

	go app.Run()
	time.Sleep(2 * time.Second)
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
	v.validate()
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
	err = app.RegisterBean(&order2{t: t}, bean.SetOrder(2))
	if err != nil {
		t.Fatal(err)
	}
	err = app.RegisterBeanByName("c", &order3{t: t}, bean.SetOrder(3))
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
	ret := &customerEvent{}
	ret.ResetOccurredTime()
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
	event.GetContext().PublishEvent(newCustomerEvent("hello world"))
	event.GetContext().PublishEvent(appcontext.NewPayloadApplicationEvent("hello world"))
	event.GetContext().PublishEvent(appcontext.NewPayloadApplicationEvent(&aImpl{v: "hello world2"}))
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

func (l *listener3) RegisterConsumer(register appcontext.ApplicationEventConsumerRegister) error {
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
