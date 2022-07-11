/*
 * Copyright (C) 2022, Xiongfa Li.
 * All rights reserved.
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

package errors

import "strings"

type Errors []error

func (es Errors) Empty() bool {
	return len(es) == 0
}

func (es *Errors) AddError(e error) *Errors {
	*es = append(*es, e)
	return es
}

func (es Errors) Error() string {
	buf := strings.Builder{}
	for i := range es {
		buf.WriteString(es[i].Error())
		if i < len(es)-1 {
			buf.WriteString(",")
		}
	}
	return buf.String()
}
