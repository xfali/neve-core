// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package injector

import (
	"errors"
	"github.com/xfali/neve-core/bean"
	reflection2 "github.com/xfali/neve-core/reflection"
	"github.com/xfali/neve-utils/reflection"
	"github.com/xfali/xlog"
	"reflect"
)

const (
	injectTagName = "inject"
)

type Injector interface {
	Inject(container bean.Container, o interface{}) error
}

type Actuator func(c bean.Container, name string, v reflect.Value) error

type defaultInjector struct {
	logger    xlog.Logger
	actuators map[reflect.Kind]Actuator
	recursive bool
}

type Opt func(*defaultInjector)

func New(opts ...Opt) *defaultInjector {
	ret := &defaultInjector{
		logger: xlog.GetLogger(),
	}
	ret.actuators = map[reflect.Kind]Actuator{
		reflect.Interface: ret.injectInterface,
		reflect.Struct:    ret.injectStruct,
		reflect.Slice:     ret.injectSlice,
		reflect.Map:       ret.injectMap,
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func (injector *defaultInjector) Inject(c bean.Container, o interface{}) error {
	v := reflect.ValueOf(o)
	if v.Kind() == reflect.Interface {
		return injector.injectInterface(c, "", v)
	}
	t := v.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	if t.Kind() == reflect.Struct {
		return injector.injectStructFields(c, v)
	}
	return errors.New("Type Not support. ")
}

func (injector *defaultInjector) injectStructFields(c bean.Container, v reflect.Value) error {
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return errors.New("result must be struct ptr")
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag, ok := field.Tag.Lookup(injectTagName)
		if ok {
			fieldValue := v.Field(i)
			fieldType := fieldValue.Type()
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}
			if fieldValue.CanSet() {
				actuate := injector.actuators[fieldType.Kind()]
				if actuate == nil {
					return errors.New("Cannot inject this kind: " + fieldType.Name())
				}
				err := actuate(c, tag, fieldValue)
				if err != nil {
					injector.logger.Errorf("Inject Field error: [%s: %s] %s\n ", reflection.GetTypeName(t), field.Name, err.Error())
				}
			} else {
				injector.logger.Errorf("Inject failed: Field cannot SET [%s: %s]\n ", reflection.GetTypeName(t), field.Name)
			}
		}
	}

	return nil
}

func (injector *defaultInjector) injectInterface(c bean.Container, name string, v reflect.Value) error {
	vt := v.Type()
	if name == "" {
		name = reflection.GetTypeName(vt)
	}
	o, ok := c.GetDefinition(name)
	if ok {
		v.Set(o.Value())
		return nil
	} else {
		//自动注入
		var matchValues []bean.Definition
		c.Scan(func(key string, value bean.Definition) bool {
			//指定名称注册的对象直接跳过，因为在container.Get未满足，所以认定不是用户想要注入的对象
			if key != value.Name() {
				return true
			}
			ot := value.Type()
			if ot.Implements(vt) {
				matchValues = append(matchValues, value)
				if len(matchValues) > 1 {
					panic("Auto Inject bean found more than 1")
				}
				return true
			}
			return true
		})
		if len(matchValues) == 1 {
			v.Set(matchValues[0].Value())
			// cache to container
			err := c.PutDefinition(reflection.GetTypeName(vt), matchValues[0])
			if err != nil {
				injector.logger.Warnln(err)
			}
			return nil
		}
	}
	return errors.New("Inject nothing, cannot find any Implementation: " + reflection.GetTypeName(vt))
}

func (injector *defaultInjector) injectSlice(c bean.Container, name string, v reflect.Value) error {
	vt := v.Type()
	if name == "" {
		name = reflection2.GetSliceName(vt)
	}
	elemType := vt.Elem()
	o, ok := c.GetDefinition(name)
	if ok {
		return reflection2.SmartCopySlice(v, o.Value())
	} else {
		//自动注入
		destTmp := sliceAppender{
			v:        v,
			elemType: elemType,
		}
		c.Scan(destTmp.Scan)
		destTmp.Set(v)
		if v.Len() > 0 {
			// cache to container
			bean, err := bean.CreateBeanDefinition(v.Interface())
			if err != nil {
				injector.logger.Warnln(err)
			}
			err = c.PutDefinition(reflection2.GetSliceName(vt), bean)
			if err != nil {
				injector.logger.Warnln(err)
			}
			return nil
		}
	}
	return errors.New("Inject nothing, cannot find any Implementation: " + reflection2.GetSliceName(vt))
}

func (injector *defaultInjector) injectMap(c bean.Container, name string, v reflect.Value) error {
	vt := v.Type()
	if name == "" {
		name = reflection2.GetMapName(vt)
	}
	keyType := vt.Key()
	elemType := vt.Elem()
	o, ok := c.GetDefinition(name)
	if ok {
		v.Set(o.Value())
		return nil
	} else {
		if keyType.Kind() != reflect.String {
			return errors.New("Key type must be string. ")
		}
		//自动注入
		c.Scan(func(key string, value bean.Definition) bool {
			ot := value.Type()
			// interface
			if elemType.Kind() == reflect.Interface {
				if ot.Implements(elemType) {
					v.SetMapIndex(reflect.ValueOf(key), value.Value())
				}
			} else if elemType.Kind() == reflect.Ptr {
				if elemType == value.Type() {
					v.SetMapIndex(reflect.ValueOf(key), value.Value())
				} else if ot.ConvertibleTo(elemType) {
					v.SetMapIndex(reflect.ValueOf(key), value.Value().Convert(elemType))
				}
			}

			return true
		})
		if v.Len() > 0 {
			// cache to container
			bean, err := bean.CreateBeanDefinition(v.Interface())
			if err != nil {
				injector.logger.Warnln(err)
			}
			err = c.PutDefinition(reflection2.GetMapName(vt), bean)
			if err != nil {
				injector.logger.Warnln(err)
			}
			return nil
		}
	}
	return errors.New("Inject nothing, cannot find any Implementation: " + reflection2.GetMapName(vt))
}

func (injector *defaultInjector) injectStruct(c bean.Container, name string, v reflect.Value) error {
	vt := v.Type()
	if name == "" {
		name = reflection.GetTypeName(vt)
	}
	o, ok := c.GetDefinition(name)
	if ok {
		ov := o.Value()
		if vt.Kind() == reflect.Ptr {
			v.Set(ov)
		} else {
			// 只允许注入指针类型
			injector.logger.Errorf("Inject struct: [%s] failed: field must be pointer. ", reflection.GetTypeName(vt))
			//v.Set(ov.Elem())
		}
		return nil
	}

	if injector.recursive {
		return injector.injectStructFields(c, v)
	} else {
		return errors.New("Inject nothing, cannot find any instance of  " + reflection.GetTypeName(vt))
	}
}

func OptSetLogger(v xlog.Logger) Opt {
	return func(injector *defaultInjector) {
		injector.logger = v
	}
}

func OptSetRecursive(recursive bool) Opt {
	return func(injector *defaultInjector) {
		injector.recursive = recursive
	}
}

// 配置注入执行器，使其能注入更多类型
func OptSetActuator(kind reflect.Kind, actuator Actuator) Opt {
	return func(injector *defaultInjector) {
		if actuator != nil {
			injector.actuators[kind] = actuator
		}
	}
}

type sliceAppender struct {
	v        reflect.Value
	elemType reflect.Type
}

func (s *sliceAppender) Set(value reflect.Value) error {
	value.Set(s.v)
	return nil
}

func (s *sliceAppender) Scan(key string, value bean.Definition) bool {
	ot := value.Type()
	// interface
	if s.elemType.Kind() == reflect.Interface {
		if ot.Implements(s.elemType) {
			s.v = reflect.Append(s.v, value.Value())
		}
	} else if s.elemType.Kind() == reflect.Ptr {
		if s.elemType == value.Type() {
			s.v = reflect.Append(s.v, value.Value())
		} else if ot.ConvertibleTo(s.elemType) {
			s.v = reflect.Append(s.v, value.Value().Convert(s.elemType))
		}
	}

	return true
}
