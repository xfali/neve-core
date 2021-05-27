// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

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
