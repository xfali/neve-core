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

package bean

import (
	"fmt"
	"reflect"
)

type Classifier interface {
	// 对象分类，判断对象是否实现某些接口，并进行相关归类。为了支持多协程处理，该方法应线程安全。
	// 注意：该方法建议只做归类，具体处理使用Process，不保证Processor的实现在此方法中做了相关处理。
	// 该方法在Bean Inject注入之后调用
	// return: bool 是否能够处理对象， error 处理是否有错误
	Classify(o interface{}) (bool, error)
}

type Definition interface {
	// 类型
	Type() reflect.Type

	// 名称
	Name() string

	// 获得值
	Value() reflect.Value

	// 获得注册的bean对象
	Interface() interface{}

	// 是否是可注入对象
	IsObject() bool

	// 在属性配置完成后调用
	AfterSet() error

	// 销毁对象
	Destroy() error

	// 对对象进行分类
	Classify(classifier Classifier) (bool, error)
}

type DefinitionCreator func(o interface{}) (Definition, error)

var (
	InitializingType       = reflect.TypeOf((*Initializing)(nil)).Elem()
	DisposableType         = reflect.TypeOf((*Disposable)(nil)).Elem()
	ErrorType              = reflect.TypeOf((*error)(nil)).Elem()
	beanDefinitionCreators = map[reflect.Kind]DefinitionCreator{
		reflect.Ptr:   newObjectDefinition,
		reflect.Func:  newFunctionExDefinition,
		reflect.Slice: newSliceDefinition,
		reflect.Map:   newMapDefinition,
	}
)

// 注册BeanDefinition创建器，使其能处理更多类型。
// 默认支持Pointer、Function
func RegisterBeanDefinitionCreator(kind reflect.Kind, creator DefinitionCreator) {
	if creator != nil {
		beanDefinitionCreators[kind] = creator
	}
}

func CreateBeanDefinition(o interface{}) (Definition, error) {
	if v, ok := o.(CustomBeanFactory); ok {
		return newCustomMethodBeanDefinition(v)
	}

	t := reflect.TypeOf(o)
	creator, ok := beanDefinitionCreators[t.Kind()]
	if !ok || creator == nil {
		return nil, fmt.Errorf("Cannot handle this type: %s" + t.String())
	}

	return creator(o)
}
