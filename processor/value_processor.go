// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package processor

import (
	"github.com/xfali/fig"
	"github.com/xfali/neve-core/container"
)

type ValueProcessor struct {
	conf      fig.Properties
	tagPxName string
	tagName   string
}

type Opt func(processor *ValueProcessor)

func OptSetValueTag(tagPxName, tagName string) Opt {
	return func(processor *ValueProcessor) {
		if tagName != "" {
			if tagPxName == "" {
				tagPxName = fig.TagPrefixName
			}
			processor.tagName = tagName
			processor.tagPxName = tagPxName
		}
	}
}

func NewValueProcessor(opts ...Opt) *ValueProcessor {
	ret := &ValueProcessor{
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func (p *ValueProcessor) Init(conf fig.Properties, container container.Container) error {
	p.conf = conf
	return nil
}

func (p *ValueProcessor) Classify(o interface{}) (bool, error) {
	if p.tagName == "" {
		return true, fig.Fill(p.conf, o)
	} else {
		return true, fig.FillExWithTagName(p.conf, o, false, p.tagPxName, p.tagName)
	}
}

func (p *ValueProcessor) Process() error {
	return nil
}

func (p *ValueProcessor) Destroy() error {
	return nil
}
