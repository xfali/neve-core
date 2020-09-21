// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

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
			xlog.Infof("get a signal %s, stop the server", si.String())
			go func() {
				for i := range closers {
					cErr := closers[i]()
					if cErr != nil {
						logger.Errorln(cErr)
						err = cErr
					}
				}
			}()
			time.Sleep(2 * time.Second)
			xlog.Infof("------ Process exited ------")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
