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

package reflection

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func GetTypeName(t reflect.Type) string {
	buf := strings.Builder{}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		buf.WriteString("*")
	}

	switch t.Kind() {
	case reflect.Slice:
		buf.WriteString(GetSliceName(t))
		break
	case reflect.Map:
		buf.WriteString(GetMapName(t))
		break
	default:
		name := t.PkgPath()
		if name != "" {
			buf.WriteString(strings.Replace(name, "/", ".", -1) + "." + t.Name())
			break
		} else {
			buf.WriteString(t.Name())
			break
		}
	}
	return buf.String()
}

func GetSliceName(t reflect.Type) string {
	elemType := t.Elem()

	name := elemType.PkgPath()
	if name != "" {
		name = strings.Replace(name, "/", ".", -1) + "." + elemType.Name()
		return "[]" + name
	} else {
		return t.String()
	}
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
		return fmt.Sprintf("map[%s]%s", key, name)
	} else {
		return t.String()
	}
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
				if ot.AssignableTo(destElemType) {
					destTmp = reflect.Append(destTmp, value)
				}
				//if destElemType.Kind() == reflect.Interface {
				//	if ot.Implements(destElemType) {
				//		destTmp = reflect.Append(destTmp, value)
				//	}
				//} else if destElemType.Kind() == reflect.Ptr {
				//	if destElemType == value.Type() || ot.AssignableTo(destElemType) {
				//		destTmp = reflect.Append(destTmp, value)
				//	}
				//	//else if ot.ConvertibleTo(destElemType) {
				//	//	destTmp = reflect.Append(destTmp, value.Convert(destElemType))
				//	//}
				//} else {
				//	//return errors.New("Type not match. ")
				//}
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
				if ot.AssignableTo(destElemType) {
					dest.SetMapIndex(key, value)
				}
				//if destElemType.Kind() == reflect.Interface {
				//	if ot.Implements(destElemType) {
				//		dest.SetMapIndex(key, value)
				//	}
				//} else if destElemType.Kind() == reflect.Ptr {
				//	if destElemType == value.Type() || ot.AssignableTo(destElemType) {
				//		dest.SetMapIndex(key, value)
				//	}
				//} else {
				//	//return errors.New("Type not match. ")
				//}
			}
			dest.Set(destTmp)
		}
	}
	return nil
}
