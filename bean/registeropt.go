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

const (
	KeySetOrder = "register.bean.order"
)

type Setter interface {
	Set(key string, value interface{})
}

// Bean注册配置，已支持的配置有：
// * bean.SetOrder(int) 配置bean注入顺序
type RegisterOpt func(setter Setter)

// 配置bean注入顺序
func SetOrder(order int) RegisterOpt {
	return func(setter Setter) {
		setter.Set(KeySetOrder, order)
	}
}
