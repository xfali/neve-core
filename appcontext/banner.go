// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package appcontext

import (
	"bytes"
	"github.com/xfali/xlog"
	"io"
	"os"
)

const (
	neveBanner = `
  .\'/.   .-----.-----.--.--.-----.
->- x -<- |     |  -__|  |  |  -__|
  '/.\'   |__|__|_____|\___/|_____|
=================  (v0.1.1.RELEASE)
`
)

func printBanner(bannerPath string) {
	output := []byte(neveBanner)
	if bannerPath != "" {
		f, err := os.Open(bannerPath)
		if err == nil {
			buf := bytes.NewBuffer(nil)
			_, err := io.Copy(buf, f)
			if err == nil {
				output = buf.Bytes()
			}
		}
	}
	selectWriter().Write(output)
}

func selectWriter() io.Writer {
	for i := xlog.INFO; i <= xlog.DEBUG; i++ {
		w := xlog.GetOutputBySeverity(i)
		if w != nil {
			return w
		}
	}
	return os.Stdout
}
