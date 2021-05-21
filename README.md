# neve-core

neve-core是neve的核心组件，实现基于反射的依赖注入框架以及注册bean的生命周期管理，同时提供一套事件处理框架。

## 安装
```
go get github.com/xfali/neve-core
```

## 使用
  
### 1. 初始化
目前支持配置文件初始化：
```
app := neve.NewFileConfigApplication("assets/application-test.yaml")
```

### 2. 配置
在application-test.yaml中配置示例如下：
```
neve:
  application:
    name: Neve test application
    banner: "banner.txt"
    bannerMode: off
  inject:
    disable: false
    workers: 1

userdata:
  value: "this is a test"
  gopath: {{ env "GOPATH" }}
```
* 【neve.application.name】应用名称
* 【neve.application.banner】banner文件路径
* 【neve.application.bannerMode】如果设置为off则关闭显示banner
* 【neve.inject.disable】是否关闭注入功能，默认false，即开启依赖注入
* 【neve.inject.workers】并行注入的任务数，目前还未开放故默认为1
* 【userdata】非内置配置属性，属于用户自定义的value，可自定义名称
* 配置可使用{{ env "ENV_NAME" DEFAULT_VALUE }}或{{.Env.ENV_NAME}}获取环境变量的值，在读取时进行替换(规则见[fig](https://github.com/xfali/fig))。

### 3. 注册
#### 3.1 快速入门
* 直接注册：RegisterBean
* 按名称注册：RegisterBeanByName
```
app.RegisterBean(processor.NewValueProcessor())
app.RegisterBean(&aImpl{v: "0"})
app.RegisterBeanByName("b", &bImpl{})

// 可以用下述方式快速抛出错误
neverror.PanicError(app.RegisterBean(&aImpl{v: "0"}))
neverror.PanicError(app.RegisterBeanByName("b", &bImpl{}))
```
#### 3.2 注册参数
neve在注册时可以添加配置参数，目前支持的配置参数有
* order：影响注入的顺序（按从小到大排序，默认为0），同时也影响调用BeanAfterSet回调的顺序。
```
app.RegisterBean(NewBean(), bean.SetOrder(2))
```

### 4. 注入
#### 4.1 使用tag：inject
注意注入的field首字母需大写（Public）
* 当inject的value为空时自动选择注入
* 当inject的value不为空时则按名称注入
```
type injectBean struct {
	A  a      `inject:""`
	B  a      `inject:"b"`
	BS *bImpl `inject:"b"`
	Bf a      `inject:"c"`
}
```
#### 4.2 注入类型支持：
* interface（接口）：neve会自动选择实现对象注入或按指定名称注入
* struct Pointer（结构体指针）：neve选择具体注册的结构体对象进行注入
* slice：neve可按名称注入slice，也可以自动查询所有适配的对象添加到slice中
* map：neve可按名称注入map，也可以自动查询所有适配的对象添加到map中

#### 4.3 自定义tag：
```
app := neve.NewFileConfigApplication("assets/application-test.yaml",
			neve.OptSetInjectTagName("Autowired"))
type injectBeanB struct {
	A  a      `Autowired:""`
	B  a      `Autowired:"b"`
	BS *bImpl `Autowired:"b"`
	Bf a      `Autowired:"c"`
}
```

### 5. 注意事项
1. 注入struct Pointer时，名称必须完全匹配：
* 如注册时使用RegisterBeanByName方法，则inject的tag value必须与注册时的name完全匹配，否则无法注入。
* 如注册时使用RegisterBean方法，则inject的tag value必须为空，否则无法注入。
1. 注入interface时：
* 如注册时使用RegisterBeanByName方法，则inject的tag value必须与注册时的name完全匹配，否则无法注入。
* 如注册时使用RegisterBean方法，则inject的tag value应为空，neve会自动匹配实现interface的对象进行注入。
* 如注册时使用RegisterBeanByName方法且inject的tag value为空，则不会注入该对象。

### 6. 处理器
Processor为注册对象处理器，可对注册对象进行分类和处理。

具体参见[Processor](processor/processor.go)

内置的Processor有
* [ValueProcessor](processor/value_processor.go)（值注入处理器），负责将配置文件中的属性值注入到对象：

注意注入的field首字母需大写（Public）
```
app.RegisterBean(processor.NewValueProcessor())
app.RegisterBean(&bImpl{})

type bImpl struct {
	V string `fig:"userdata.value"`
}
```
注入的tag默认为fig，可以通过下述方法修改：
```
// 将tag修改为value
app.RegisterBean(processor.NewValueProcessor(processor.OptSetValueTag("", "value")))
err = app.RegisterBean(&bImpl{})
type bImpl struct {
	V string `value:"userdata.value"`
}
```
* [neve-web](https://github.com/xfali/neve-web) neve的WEB扩展组件，用于集成WEB相关服务。
* [neve-database](https://github.com/xfali/neve-database) neve的数据库扩展组件，用于集成数据库相关操作。

### 7. Bean生命周期
注册的对象可以实现下述方法监听容器中的bean生命周期：
```
type Initializing interface {
	// 当初始化和注入完成时回调
	BeanAfterSet() error
}

type Disposable interface {
	// 进入销毁阶段，应该尽快做回收处理并退出处理任务
	BeanDestroy() error
}
```
* BeanAfterSet (仅调用一次)：

  在bean完成对象及值注入之后，Processor的Process之前调用，可以做一些初始化相关的工作，此时从配置文件中的Value已经可以读取。
* BeanDestroy (仅调用一次)：

  在Application即将退出时调用。

### 8. 获得ApplicationContext
实现SetApplicationContext(ctx ApplicationContext)方法，在bean注入之前即可获取ApplicationContext的引用
```
type aware struct {
	t *testing.T
}

func (a *aware) SetApplicationContext(ctx appcontext.ApplicationContext) {
	if ctx == nil {
		a.t.Fatal("must not be nil")
	}
	a.t.Log(ctx.GetApplicationName())
}
```
### 9. ApplicationEvent
neve提供了事件处理框架，包含事件发布器及事件监听器。

#### 9.1 ApplicationEventPublisher
neve通过ApplicationEventPublisher发布事件，常用的事件发布方法有两种

##### 9.1.1  ApplicationContext的PublishEvent
获得ApplicationContext，通过获得ApplicationContext的PublishEvent方法发布事件
```
appCtx.PublishEvent(newCustomerEvent("hello world"))
```

##### 9.1.2  注入
neve默认会把内置的事件发布器注册到bean容器中，可以通过注入的方式获得ApplicationEventPublisher：
```
type TestBean struct {
	Publisher appcontext.ApplicationEventPublisher `inject:""`
}

// 发布事件
testBean.Publisher.PublishEvent(newCustomerEvent("hello world"))
```

##### 9.2  ApplicationEventListener
有如下多种方式注册事件监听器

##### 9.2.1  通过Application / ApplicationContext的AddListeners方法：

```
app.AddListeners(o)

appCtx.AddListeners(o)
```
该方法支持注册
* 实现ApplicationEventListener的对象
* 方法，该方法的参数为ApplicationEvent或实现ApplicationEvent的对象

##### 9.2.2 通过实现ApplicationEventConsumer接口注册消费事件的方法
```
type listener3 struct {
	t *testing.T
}

// 实现ApplicationEventConsumer接口
func (l *listener3) RegisterConsumer(register appcontext.ApplicationEventConsumerRegister) error{
	return register.RegisterApplicationEventConsumer(l.handlerEvent)
}

// 当事件为*customerEvent类型时自动匹配并调用该方法
// customerEvent需实现ApplicationEvent接口
func (l *listener3) handlerEvent(event *customerEvent) {
	l.t.Log("listener3", event.payload)
}
```
##### 9.2.3 PayloadEventListener

step 1: 定义一个类型，包含一个获取payload的方法，如getPayload
```
func (l *listener) getPayload(payload a) {
	if payload.Get() != "hello world2" {
		l.t.Fatal("not match")
	}
	l.t.Log("listener", payload.Get())
}
```
step 2: 注册PayloadEventListener
```
app.RegisterBean(appcontext.NewPayloadEventListener(l.getPayload))
```
step 3: 发布一个PayloadApplicationEvent，参数中的对象会被自动匹配并调用PayloadEventListener关联的方法
```
appContext.PublishEvent(appcontext.NewPayloadApplicationEvent(&aImpl{v: "hello world2"}))
```

### 10. 多例
neve注册和注入默认为单例，可以通过注册func() TYPE函数的方式，选择返回单例或者多例。
```
app.RegisterBean(func() a {
    return &bImpl{V: "hello world"}
})
app.RegisterBeanByName("c", func() a {
    return &bImpl{V: "hello world"}
})
```
注意：目前注册函数的方式返回的值不能在返回时自动注入对象和配置的值，请谨慎使用。
