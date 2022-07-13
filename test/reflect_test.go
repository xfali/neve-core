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
	"reflect"
	"testing"
)

func TestSet(t *testing.T) {
	t.Run("i", func(t *testing.T) {
		c := 10
		cp := &c

		src := reflect.ValueOf(cp).Elem()
		//dest := reflect.New(reflect.TypeOf(cp)).Elem()
		var i int
		dest := reflect.ValueOf(&i).Elem()
		dest.Set(src)
		t.Log(dest.Interface())
	})
	t.Run("new", func(t *testing.T) {
		c := 10
		cp := &c

		src := reflect.ValueOf(cp)
		dest := reflect.New(reflect.TypeOf(cp)).Elem()
		dest.Set(src)
		t.Log(dest.Elem().Interface())
	})

	t.Run("new", func(t *testing.T) {
		c := 10
		cp := &c

		src := reflect.ValueOf(cp)
		dest := reflect.New(reflect.TypeOf(cp)).Elem()
		dest.Set(src)
		t.Log(dest.Elem().Interface())
	})
}
