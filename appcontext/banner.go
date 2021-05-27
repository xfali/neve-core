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
	neveBanner =
` .\'/. .-----.-----.--.--.-----.
-- * --|     |  -__|  |  |  -__|
 '/.\' |__|__|_____|\___/|_____|
`
	neveBannerVersion = `================================`
)

func printNeveInfo(version string, bannerPath string, banner bool) {
	w := selectWriter()
	buf := bytes.NewBuffer(nil)
	buf.Grow(len(neveBanner) + len(neveBannerVersion) + 2)
	buf.WriteByte('\n')
	if banner {
		buf.WriteString(bannerString(bannerPath))
	}
	buf.WriteString(versionString(version, banner))
	buf.WriteByte('\n')

	w.Write(buf.Bytes())
}

func bannerString(bannerPath string) string {
	output := []byte(neveBanner)
	if bannerPath != "" {
		f, err := os.Open(bannerPath)
		if err == nil {
			buf := bytes.NewBuffer(nil)
			_, err := io.Copy(buf, f)
			if err == nil {
				if buf.Bytes()[buf.Len()-1] != '\n' {
					buf.WriteByte('\n')
				}
				output = buf.Bytes()
			}
		}
	}
	return string(output)
}

func versionString(version string, banner bool) string {
	if banner {
		size := len(version)
		bs := len(neveBannerVersion)
		if size == 0 || size > bs-3 {
			return neveBannerVersion
		}
		return fmt.Sprintf("%s (%s)\n", neveBannerVersion[:bs-size-3], version)
	} else {
		return fmt.Sprintf("=== neve === (%s)\n", version)
	}
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
