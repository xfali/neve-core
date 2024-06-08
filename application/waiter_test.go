/*
 * Copyright (C) 2024, Xiongfa Li.
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

package application

import (
	"context"
	"syscall"
	"testing"
	"time"
)

func TestSignalWaiter(t *testing.T) {
	waiter := NewSignalWaiter()
	t.Run("stop", func(t *testing.T) {
		go func() {
			time.Sleep(5 * time.Second)
			waiter.Stop()
		}()
		err := waiter.Wait(context.Background())
		t.Log(err)
	})
	t.Run("notify SIGQUIT", func(t *testing.T) {
		go func() {
			time.Sleep(5 * time.Second)
			waiter.Notify(syscall.SIGQUIT)
		}()
		err := waiter.Wait(context.Background())
		t.Log(err)
	})

	t.Run("notify SIGHUP and SIGTERM", func(t *testing.T) {
		go func() {
			time.Sleep(2 * time.Second)
			waiter.Notify(syscall.SIGHUP)

			time.Sleep(2 * time.Second)
			waiter.Notify(syscall.SIGTERM)
		}()
		err := waiter.Wait(context.Background())
		t.Log(err)
	})
}
