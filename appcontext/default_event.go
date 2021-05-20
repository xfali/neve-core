// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package appcontext

import (
	"errors"
	"github.com/xfali/xlog"
	"reflect"
	"sync"
)

const (
	defaultEventBufferSize = 4096
)

var eventType = reflect.TypeOf((*ApplicationEvent)(nil)).Elem()

type defaultEventProcessor struct {
	logger xlog.Logger

	listeners    []ApplicationEventListener
	listenerLock sync.Mutex

	eventBufSize int
	eventChan    chan ApplicationEvent

	stopChan   chan struct{}
	finishChan chan struct{}
}

type EventProcessorOpt func(processor *defaultEventProcessor)

func NewEventProcessor(opts ...EventProcessorOpt) *defaultEventProcessor {
	ret := &defaultEventProcessor{
		logger:       xlog.GetLogger(),
		eventBufSize: defaultEventBufferSize,
	}

	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func OptSetEventProcessorLogger(logger xlog.Logger) EventProcessorOpt {
	return func(proc *defaultEventProcessor) {
		proc.logger = logger
	}
}

// set event channel buffer size
func OptSetEventBufferSize(size int) EventProcessorOpt {
	return func(proc *defaultEventProcessor) {
		proc.eventBufSize = size
	}
}

func (h *defaultEventProcessor) Start() error {
	h.eventChan = make(chan ApplicationEvent, h.eventBufSize)
	h.stopChan = make(chan struct{})
	h.finishChan = make(chan struct{})

	go h.eventLoop()

	return nil
}

func (h *defaultEventProcessor) addListener(l ApplicationEventListener, withLock bool) {
	if !withLock {
		h.listeners = append(h.listeners, l)
		return
	}

	h.listenerLock.Lock()
	defer h.listenerLock.Unlock()

	h.listeners = append(h.listeners, l)
}

func (h *defaultEventProcessor) processListener(o interface{}) {
	l := h.classifyListenerInterface(o)
	if l != nil {
		h.addListener(l, true)
		return
	}

	l, err := parseListener(o)
	if err != nil {
		//ctx.logger.Errorln(err)
	} else if l != nil {
		h.addListener(l, true)
	}
}

func (h *defaultEventProcessor) classifyListenerInterface(o interface{}) ApplicationEventListener {
	if l, ok := o.(ApplicationEventListener); ok {
		return l
	}

	if c, ok := o.(ApplicationEventConsumer); ok {
		l, err := parseListener(c.GetApplicationEventConsumer())
		if err != nil {
			h.logger.Errorln(err)
			return nil
		}
		return l
	}
	return nil
}

func (h *defaultEventProcessor) Close() (err error) {
	close(h.stopChan)
	//wait for eventLoop exit
	<-h.finishChan
	return
}

func (h *defaultEventProcessor) notifyEvent(e ApplicationEvent) {
	h.listenerLock.Lock()
	defer h.listenerLock.Unlock()

	for _, v := range h.listeners {
		v.OnApplicationEvent(e)
	}
}

func (h *defaultEventProcessor) eventLoop() {
	defer close(h.finishChan)
	for {
		select {
		case <-h.stopChan:
			size := len(h.eventChan)
			for i := 0; i < size; i++ {
				h.notifyEvent(<-h.eventChan)
			}
			return
		case e, ok := <-h.eventChan:
			if ok {
				h.notifyEvent(e)
			}
		}
	}
}

func (h *defaultEventProcessor) AddListeners(listeners ...interface{}) {
	for _, o := range listeners {
		h.processListener(o)
	}
}

func (h *defaultEventProcessor) PublishEvent(e ApplicationEvent) error {
	if e == nil {
		return errors.New("event is nil. ")
	}
	select {
	case h.eventChan <- e:
		return nil
	default:
		return errors.New("event queue is full. ")
	}
}

func (h *defaultEventProcessor) NotifyEvent(e ApplicationEvent) error {
	h.notifyEvent(e)
	return nil
}

type eventProcessor struct {
	et reflect.Type
	fv reflect.Value
}

func parseListener(o interface{}) (ApplicationEventListener, error) {
	t := reflect.TypeOf(o)
	if t.Kind() != reflect.Func {
		return nil, errors.New("Param is not a function. ")
	}

	if t.NumIn() != 1 {
		return nil, errors.New("Param is not match, expect func(ApplicationEvent). ")
	}

	et := t.In(0)
	if !et.Implements(eventType) {
		return nil, errors.New("Param is not match, function param must Implements ApplicationEvent. ")
	}

	return &eventProcessor{
		et: et,
		fv: reflect.ValueOf(o),
	}, nil
}

func (ep *eventProcessor) OnApplicationEvent(e ApplicationEvent) {
	t := reflect.TypeOf(e)
	//if t == ep.et {
	//	ep.fv.Call([]reflect.Value{reflect.ValueOf(e)})
	//}
	if t.AssignableTo(ep.et) {
		ep.fv.Call([]reflect.Value{reflect.ValueOf(e)})
	}
}

type PayloadListener struct {
	et reflect.Type
	fv reflect.Value
}

// o:获得payload的consumer 类型func(Type)
func NewPayloadListener(o interface{}) *PayloadListener {
	ret := &PayloadListener{
	}
	err := ret.RefreshPayloadHandler(o)
	if err != nil {
		panic(err)
	}
	return ret
}

func (l *PayloadListener) RefreshPayloadHandler(o interface{}) error {
	t := reflect.TypeOf(o)
	if t.Kind() != reflect.Func {
		return errors.New("Param is not a function. ")
	}

	if t.NumIn() != 1 {
		return errors.New("Param is not match, expect func(ApplicationEvent). ")
	}

	et := t.In(0)

	l.et = et
	l.fv = reflect.ValueOf(o)

	return nil
}

type PayloadApplicationEvent struct {
	BaseApplicationEvent
	payload interface{}
}

func NewPayloadApplicationEvent(payload interface{}) *PayloadApplicationEvent {
	if payload == nil {
		return nil
	}
	e := PayloadApplicationEvent{}
	e.ResetOccurredTime()
	e.payload = payload
	return &e
}

func (l *PayloadListener) OnApplicationEvent(e ApplicationEvent) {
	if l.fv.IsValid() {
		if pe, ok := e.(*PayloadApplicationEvent); ok {
			payload := pe.payload
			t := reflect.TypeOf(payload)
			if t.AssignableTo(l.et) {
				l.fv.Call([]reflect.Value{reflect.ValueOf(payload)})
			}
		}
	}
}
