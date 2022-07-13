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
	"errors"
	"github.com/xfali/neve-core/reflection"
	"reflect"
	"sync/atomic"
)

type objectDefinition struct {
	name        string
	o           interface{}
	t           reflect.Type
	flagSet     int32
	flagDestroy int32
}

func newObjectDefinition(o interface{}) (Definition, error) {
	t := reflect.TypeOf(o)
	if t.Kind() == reflect.Ptr {
		t2 := t.Elem()
		if t2.Kind() == reflect.Ptr {
			return nil, errors.New("o must be a Pointer but get Pointer's Pointer")
		}
	}
	return &objectDefinition{
		name:        reflection.GetTypeName(t),
		o:           o,
		t:           t,
		flagSet:     0,
		flagDestroy: 0,
	}, nil
}

func (d *objectDefinition) Type() reflect.Type {
	return d.t
}

func (d *objectDefinition) Name() string {
	return d.name
}

func (d *objectDefinition) Value() reflect.Value {
	return reflect.ValueOf(d.o)
}

func (d *objectDefinition) Interface() interface{} {
	return d.o
}

func (d *objectDefinition) IsObject() bool {
	return true
}

func (d *objectDefinition) AfterSet() error {
	// Just run once
	if atomic.CompareAndSwapInt32(&d.flagSet, 0, 1) {
		if v, ok := d.o.(Initializing); ok {
			return v.BeanAfterSet()
		}
	}
	return nil
}

func (d *objectDefinition) Destroy() error {
	// Just run once
	if atomic.CompareAndSwapInt32(&d.flagDestroy, 0, 1) {
		if v, ok := d.o.(Disposable); ok {
			return v.BeanDestroy()
		}
	}
	return nil
}

func (d *objectDefinition) Classify(classifier Classifier) (bool, error) {
	return classifier.Classify(d.o)
}