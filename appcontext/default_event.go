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

package appcontext

import (
	"context"
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

	consumerListenerFac func() ApplicationEventConsumerListener

	stopChan   chan struct{}
	finishChan chan struct{}
	closeOnce  sync.Once
}

type EventProcessorOpt func(processor *defaultEventProcessor)

func NewEventProcessor(opts ...EventProcessorOpt) *defaultEventProcessor {
	ret := &defaultEventProcessor{
		logger:              xlog.GetLogger(),
		eventBufSize:        defaultEventBufferSize,
		consumerListenerFac: defaultConsumerListenerFac,
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

func OptSetConsumerListenerFactory(fac func() ApplicationEventConsumerListener) EventProcessorOpt {
	return func(processor *defaultEventProcessor) {
		processor.consumerListenerFac = fac
	}
}

func (h *defaultEventProcessor) BeanAfterSet() error {
	return h.Start()
}

func (h *defaultEventProcessor) BeanDestroy() error {
	return h.Close()
}

func (h *defaultEventProcessor) Start() error {
	h.eventChan = make(chan ApplicationEvent, h.eventBufSize)
	h.stopChan = make(chan struct{})
	h.finishChan = make(chan struct{})
	h.closeOnce = sync.Once{}

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

	l, err := h.parseListener(o)
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
		l := h.createConsumerListener()
		err := c.RegisterConsumer(l)
		if err != nil {
			h.logger.Errorln(err)
			return nil
		}
		return l
	}
	return nil
}

func (h *defaultEventProcessor) Close() (err error) {
	h.closeOnce.Do(func() {
		close(h.stopChan)
		//wait for eventLoop exit
		<-h.finishChan
		h.logger.Infoln("Event Processor closed.")
	})

	return
}

func (h *defaultEventProcessor) notifyEvent(e ApplicationEvent) error {
	h.listenerLock.Lock()
	defer h.listenerLock.Unlock()

	for _, v := range h.listeners {
		v.OnApplicationEvent(e)
	}
	return nil
}

func (h *defaultEventProcessor) eventLoop() {
	defer func() {
		select {
		case <-h.finishChan:
			return
		default:
			close(h.finishChan)
		}
	}()
	for {
		select {
		case <-h.stopChan:
			size := len(h.eventChan)
			for i := 0; i < size; i++ {
				err := h.notifyEvent(<-h.eventChan)
				if err != nil {
					h.logger.Errorln("Event Processor event loop notify event failed: ", err)
				}
			}
			return
		case e, ok := <-h.eventChan:
			if ok {
				err := h.notifyEvent(e)
				if err != nil {
					h.logger.Errorln("Event Processor event loop notify event failed: ", err)
				}
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

func (h *defaultEventProcessor) PostEvent(ctx context.Context, e ApplicationEvent) error {
	if e == nil {
		return errors.New("event is nil. ")
	}
	select {
	case h.eventChan <- e:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *defaultEventProcessor) SendEvent(e ApplicationEvent) error {
	return h.notifyEvent(e)
}

func (h *defaultEventProcessor) NotifyEvent(e ApplicationEvent) error {
	return h.notifyEvent(e)
}

func (h *defaultEventProcessor) createConsumerListener() ApplicationEventConsumerListener {
	return h.consumerListenerFac()
}

func (h *defaultEventProcessor) parseListener(o interface{}) (ApplicationEventListener, error) {
	l := h.createConsumerListener()
	if err := l.RegisterApplicationEventConsumer(o); err != nil {
		return nil, err
	}
	return l, nil
}

type eventProcessor struct {
	invokers []ConsumerInvoker
}

func defaultConsumerListenerFac() ApplicationEventConsumerListener {
	return &eventProcessor{}
}

func (ep *eventProcessor) RegisterApplicationEventConsumer(consumer interface{}) error {
	invoker := eventInvoker{}
	if err := invoker.ResolveConsumer(consumer); err != nil {
		return err
	}
	ep.invokers = append(ep.invokers, &invoker)
	return nil
}

func (ep *eventProcessor) OnApplicationEvent(e ApplicationEvent) {
	for _, invoker := range ep.invokers {
		invoker.Invoke(e)
	}
}

type ConsumerInvoker interface {
	// 消费
	Invoke(data interface{}) bool

	// 检查consumer是否符合类型要求
	ResolveConsumer(consumer interface{}) error
}

type consumerInvoker struct {
	et reflect.Type
	fv reflect.Value
}

func (invoker *consumerInvoker) Invoke(data interface{}) bool {
	t := reflect.TypeOf(data)
	if t.AssignableTo(invoker.et) {
		invoker.fv.Call([]reflect.Value{reflect.ValueOf(data)})
		return true
	}
	return false
}

type payloadInvoker struct {
	consumerInvoker
}

func (invoker *payloadInvoker) ResolveConsumer(consumer interface{}) error {
	t := reflect.TypeOf(consumer)
	if t.Kind() != reflect.Func {
		return errors.New("Param is not a function. ")
	}

	if t.NumIn() != 1 {
		return errors.New("Param is not match, expect func(ApplicationEvent). ")
	}

	et := t.In(0)
	invoker.et = et
	invoker.fv = reflect.ValueOf(consumer)
	return nil
}

type eventInvoker struct {
	consumerInvoker
}

func (invoker *eventInvoker) ResolveConsumer(consumer interface{}) error {
	t := reflect.TypeOf(consumer)
	if t.Kind() != reflect.Func {
		return errors.New("Param is not a function. ")
	}

	if t.NumIn() != 1 {
		return errors.New("Param is not match, expect func(ApplicationEvent). ")
	}

	et := t.In(0)
	if !et.AssignableTo(eventType) {
		return errors.New("Param is not match, function param must Implements ApplicationEvent. ")
	}

	invoker.et = et
	invoker.fv = reflect.ValueOf(consumer)
	return nil
}

type PayloadEventListener struct {
	invokers []ConsumerInvoker
}

// o:获得payload的consumer 类型func(Type)
func NewPayloadEventListener(consumer ...interface{}) *PayloadEventListener {
	ret := &PayloadEventListener{
		invokers: make([]ConsumerInvoker, 0, len(consumer)),
	}
	err := ret.RefreshPayloadHandler(consumer...)
	if err != nil {
		panic(err)
	}
	return ret
}

func (l *PayloadEventListener) RefreshPayloadHandler(consumer ...interface{}) error {
	if len(consumer) == 0 {
		return errors.New("payload callers is nil")
	}
	for _, o := range consumer {
		l.RegisterApplicationEventConsumer(o)
	}

	return nil
}

func (l *PayloadEventListener) RegisterApplicationEventConsumer(consumer interface{}) error {
	invoker := payloadInvoker{}
	if err := invoker.ResolveConsumer(consumer); err != nil {
		return err
	}
	l.invokers = append(l.invokers, &invoker)
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
	e := PayloadApplicationEvent{
		BaseApplicationEvent: *NewBaseApplicationEvent(),
	}
	e.payload = payload
	return &e
}

func (l *PayloadEventListener) OnApplicationEvent(e ApplicationEvent) {
	if len(l.invokers) > 0 {
		if pe, ok := e.(*PayloadApplicationEvent); ok {
			for _, invoker := range l.invokers {
				invoker.Invoke(pe.payload)
			}
		}
	}
}

type dummyEventProc struct{}

func NewDisableEventProcessor() *dummyEventProc {
	return &dummyEventProc{}
}

func (p *dummyEventProc) NotifyEvent(e ApplicationEvent) error {
	panic("Application event process: Disabled")
}

func (p *dummyEventProc) PublishEvent(e ApplicationEvent) error {
	panic("Application event process: Disabled")
}

func (p *dummyEventProc) PostEvent(ctx context.Context, e ApplicationEvent) error {
	panic("Application event process: Disabled")
}

func (p *dummyEventProc) SendEvent(e ApplicationEvent) error {
	panic("Application event process: Disabled")
}

func (p *dummyEventProc) AddListeners(listeners ...interface{}) {
	panic("Application event process: Disabled")
}

func (p *dummyEventProc) Start() error {
	return nil
}

func (p *dummyEventProc) Close() error {
	return nil
}
