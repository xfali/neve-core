# neve-core

neve-core是neve的核心组件，实现基于反射的依赖注入框架以及注册bean的生命周期管理。

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
  inject:
    disable: false
    workers: 1

userdata:
  value: "this is a test"
  gopath: {{ env "GOPATH" }}
```
* 【neve.inject.disable】是否关闭注入功能，默认false，即开启依赖注入
* 【neve.inject.workers】并行注入的任务数，目前还未开放故默认为1
* 【userdata】非内置配置属性，属于用户自定义的value，可自定义名称
* 配置可使用{{ env "ENV_NAME" DEFAULT_VALUE }}或{{.Env.ENV_NAME}}获取环境变量的值，在读取时进行替换(规则见[fig](https://github.com/xfali/fig))。

### 3. 注册
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
  
### 8. 多例
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
