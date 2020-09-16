// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package neve

import (
	"os"
	"path/filepath"
)

const (
	envResourceDir = "ENV_RESOURCE_DIR"
)

var ResourceRoot string

func GetResource(relFilePath string) string {
	dir := os.Getenv(envResourceDir)
	if dir == "" {
		dir = ResourceRoot
	}
	return filepath.Join(dir, relFilePath)
}

func SetResourceRoot(dir string) {
	ResourceRoot = dir
}
