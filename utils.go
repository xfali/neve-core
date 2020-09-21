// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package neve

import (
	"github.com/xfali/neve-utils/log"
	"github.com/xfali/xlog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func HandlerSignal(closers ...func() error) (err error) {
	var (
		ch = make(chan os.Signal, 1)
	)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	logger := log.GetLogger()
	for {
		si := <-ch
		switch si {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			time.Sleep(time.Second * 2)
			xlog.Infof("get a signal %s, stop the server", si.String())
			for i := range closers {
				cErr := closers[i]()
				if cErr != nil {
					logger.Errorln(cErr)
					err = cErr
				}
			}
			time.Sleep(time.Second)
			xlog.Infof("------ Process exited ------")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
