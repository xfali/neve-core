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

package neve

import (
	"github.com/xfali/xlog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func HandlerSignal(logger xlog.Logger, closers ...func() error) (err error) {
	var (
		ch = make(chan os.Signal, 1)
	)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		si := <-ch
		switch si {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			time.Sleep(100 * time.Millisecond)
			xlog.Infof("Got a signal %s, closing...", si.String())
			go func() {
				for i := range closers {
					cErr := closers[i]()
					if cErr != nil {
						logger.Errorln(cErr)
						err = cErr
					}
				}
			}()
			time.Sleep(3 * time.Second)
			xlog.Infof("------ Process exited ------")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
