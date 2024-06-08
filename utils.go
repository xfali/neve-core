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
	"context"
	"github.com/xfali/neve-core/application"
	"github.com/xfali/neve-core/errors"
	"github.com/xfali/xlog"
	"time"
)

func HandlerSignal(logger xlog.Logger, closers ...func() error) (err error) {
	defer func(pErr *error) {
		*pErr = Quit(logger, 3*time.Second, closers...)
	}(&err)
	_ = application.NewSignalWaiter(application.SignalWaiterOpts.SetLogger(logger)).Wait(context.Background())
	return
}

func Quit(logger xlog.Logger, sleepTime time.Duration, closers ...func() error) error {
	errs := &errors.LockedErrors{}
	time.Sleep(100 * time.Millisecond)
	if len(closers) > 0 {
		go func() {
			for i := range closers {
				cErr := closers[i]()
				if cErr != nil {
					logger.Errorln(cErr)
					errs.AddError(cErr)
				}
			}
		}()
	}
	if sleepTime > 0 {
		time.Sleep(sleepTime)
	}
	logger.Infof("------ Process exited ------")
	if errs.Empty() {
		return nil
	}
	return errs
}
