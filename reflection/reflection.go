// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package reflection

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func GetSliceName(t reflect.Type) string {
	elemType := t.Elem()

	name := elemType.PkgPath()
	if name != "" {
		name = strings.Replace(name, "/", ".", -1) + "." + elemType.Name()
	}
	return "[]" + name
}

func GetMapName(t reflect.Type) string {
	keyType := t.Key()
	elemType := t.Elem()

	key := keyType.PkgPath()
	if key != "" {
		key = strings.Replace(key, "/", ".", -1) + "." + keyType.Name()
	}

	name := elemType.PkgPath()
	if name != "" {
		name = strings.Replace(name, "/", ".", -1) + "." + elemType.Name()
	}
	return fmt.Sprintf("map[%s]%s", key, name)
}

func SmartCopySlice(dest, src reflect.Value) error {
	destType := dest.Type()
	destElemType := destType.Elem()
	srcType := src.Type()
	if srcType.Kind() != reflect.Slice {
		return errors.New("Src Type is not a slice. " + srcType.String())
	}
	srcElemType := srcType.Elem()

	if destType.Kind() == srcType.Kind() {
		if destElemType.Kind() == srcElemType.Kind() {
			dest.Set(src)
		} else {
			destTmp := dest
			for i := 0; i < src.Len(); i++ {
				value := src.Index(i)
				ot := value.Type()
				// interface
				if destElemType.Kind() == reflect.Interface {
					if ot.Implements(destElemType) {
						destTmp = reflect.Append(destTmp, value)
					}
				} else if destElemType.Kind() == reflect.Ptr {
					if destElemType == value.Type() {
						destTmp = reflect.Append(destTmp, value)
					} else if ot.ConvertibleTo(destElemType) {
						destTmp = reflect.Append(destTmp, value.Convert(destElemType))
					}
				} else {
					return errors.New("Type not match. ")
				}
			}
			dest.Set(destTmp)
		}
	}
	return nil
}

func SmartCopyMap(dest, src reflect.Value) error {
	destType := dest.Type()
	destElemType := destType.Elem()
	destKeyType := destType.Key()
	srcType := src.Type()
	if srcType.Kind() != reflect.Map {
		return errors.New("Src Type is not a Map. " + srcType.String())
	}
	srcElemType := srcType.Elem()
	srcKeyType := srcType.Key()

	if destKeyType != srcKeyType {
		return fmt.Errorf("expect key type: %s, but get %s. ", destKeyType.String(), srcKeyType.String())
	}
	if destType.Kind() == srcType.Kind() {
		if destElemType.Kind() == srcElemType.Kind() {
			dest.Set(src)
		} else {
			destTmp := dest
			keys := src.MapKeys()
			for _, key := range keys {
				value := src.MapIndex(key)
				ot := value.Type()
				// interface
				if destElemType.Kind() == reflect.Interface {
					if ot.Implements(destElemType) {
						dest.SetMapIndex(key, value)
					}
				} else if destElemType.Kind() == reflect.Ptr {
					if destElemType == value.Type() {
						dest.SetMapIndex(key, value)
					} else if ot.ConvertibleTo(destElemType) {
						dest.SetMapIndex(key,value.Convert(destElemType))
					}
				} else {
					return errors.New("Type not match. ")
				}
			}
			dest.Set(destTmp)
		}
	}
	return nil
}
