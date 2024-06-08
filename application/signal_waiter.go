/*
 * Copyright 2024 Xiongfa Li.
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

package application

import (
	"context"
	"github.com/xfali/xlog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type SignalWaiter interface {
	// Wait 等待信号，直到获得预期的信号后退出
	// 参数 ctx: 监听ctx，如果ctx Done则同样退出
	// 返回 err: 等待信号错误
	Wait(ctx context.Context) (err error)

	// Notify 主动发送信号
	// 参数 signal: 发送的信号
	// 返回 err: 发送错误
	Notify(signal os.Signal) (err error)

	// Stop 强制结束等待
	Stop()
}

type SignalWaiterOpt func(*defaultWaiter)

type defaultWaiter struct {
	logger        xlog.Logger
	signals       []os.Signal
	exitSignals   []os.Signal
	ignoreSignals []os.Signal
	ch            chan os.Signal
	
	ctx     context.Context
	cancel  context.CancelFunc
	ctxLock sync.Mutex
}

func NewSignalWaiter(opts ...SignalWaiterOpt) *defaultWaiter {
	ret := &defaultWaiter{
		logger:        xlog.GetLogger(),
		ch:            make(chan os.Signal, 1),
		signals:       []os.Signal{syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT},
		exitSignals:   []os.Signal{syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT},
		ignoreSignals: []os.Signal{syscall.SIGHUP},
	}
	for _, opt := range opts {
		opt(ret)
	}
	signal.Notify(ret.ch, ret.signals...)
	return ret
}

func (h *defaultWaiter) Wait(ctx context.Context) error {
	h.ctxLock.Lock()
	h.ctx, h.cancel = context.WithCancel(ctx)
	h.ctxLock.Unlock()
	for {
		select {
		case <-h.ctx.Done():
			h.logger.Infof("Context done, error: %v, closing...", ctx.Err())
			return ctx.Err()
		case si := <-h.ch:
			for _, v := range h.exitSignals {
				if si == v {
					h.logger.Infof("Got a signal %s, closing...\n", si.String())
					return nil
				}
			}
			ignore := false
			for _, v := range h.ignoreSignals {
				if si == v {
					h.logger.Infof("Ignore signal %s\n", si.String())
					ignore = true
				}
			}
			if ignore {
				continue
			}
			return nil
		}
	}
}

func (h *defaultWaiter) Notify(signal os.Signal) error {
	select {
	case h.ch <- signal:
		return nil
	default:
	}
	return nil
}

func (h *defaultWaiter) Stop() {
	h.ctxLock.Lock()
	defer h.ctxLock.Unlock()
	if h.cancel != nil {
		h.cancel()
	}
}

type signalWaiterOpts struct {
}

var SignalWaiterOpts signalWaiterOpts

func (o signalWaiterOpts) SetLogger(logger xlog.Logger) SignalWaiterOpt {
	return func(wait *defaultWaiter) {
		wait.logger = logger
	}
}

func (o signalWaiterOpts) AddNotifySignals(signals ...os.Signal) SignalWaiterOpt {
	return func(wait *defaultWaiter) {
		wait.signals = append(wait.signals, signals...)
	}
}

func (o signalWaiterOpts) AddExitSignals(signals ...os.Signal) SignalWaiterOpt {
	return func(wait *defaultWaiter) {
		wait.exitSignals = append(wait.exitSignals, signals...)
	}
}

func (o signalWaiterOpts) AddIgnoreSignals(signals ...os.Signal) SignalWaiterOpt {
	return func(wait *defaultWaiter) {
		wait.ignoreSignals = append(wait.ignoreSignals, signals...)
	}
}
