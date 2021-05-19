// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package bean

type Setter interface {
	Set(key string, value interface{})
}

type RegisterOpt func(setter Setter)

const (
	KeySetOrder = "register.bean.order"
)

func SetOrder(order int) RegisterOpt {
	return func(setter Setter) {
		setter.Set(KeySetOrder, order)
	}
}
