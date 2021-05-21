// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package appcontext

import (
	"bytes"
	"fmt"
	"github.com/xfali/xlog"
	"io"
	"os"
)

const (
	neveBanner = `
  .\'/.   .-----.-----.--.--.-----.
->- x -<- |     |  -__|  |  |  -__|
  '/.\'   |__|__|_____|\___/|_____|
`
	neveBannerVersion = `===================================`
)

//=================  (v0.1.1.RELEASE)
func printBanner(version string, bannerPath string) {
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
	w := selectWriter()
	w.Write(output)
	w.Write([]byte(versionString(version)))
}

func versionString(version string) string {
	size := len(version)
	bs := len(neveBannerVersion)
	if size == 0 || size > bs - 3{
		return neveBannerVersion
	}
	return fmt.Sprintf("%s (%s)\n", neveBannerVersion[:bs - size - 3], version)
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
