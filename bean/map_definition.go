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
	"github.com/xfali/neve-core/reflection"
	"reflect"
)

type mapDefinition struct {
	name string
	o    interface{}
	t    reflect.Type
}

func newMapDefinition(o interface{}) (Definition, error) {
	t := reflect.TypeOf(o)
	return &mapDefinition{
		name: reflection.GetMapName(t),
		o:    o,
		t:    t,
	}, nil
}

func (d *mapDefinition) Type() reflect.Type {
	return d.t
}

func (d *mapDefinition) Name() string {
	return d.name
}

func (d *mapDefinition) Value() reflect.Value {
	return reflect.ValueOf(d.o)
}

func (d *mapDefinition) Interface() interface{} {
	return d.o
}

func (d *mapDefinition) IsObject() bool {
	return false
}

func (d *mapDefinition) AfterSet() error {
	return nil
}

func (d *mapDefinition) Destroy() error {
	return nil
}

func (d *mapDefinition) Classify(classifier Classifier) (bool, error) {
	return false, nil
}
