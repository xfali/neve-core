// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

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
