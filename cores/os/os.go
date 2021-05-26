// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

package os

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Run 运行app服务器
func Run(s Server) error {
	handleSignals(s)
	if err := runService(s); err != nil {
		return err
	}

	time.Sleep(time.Second * 10)
	return nil
}

// ExecPath 获取前可执行文件完全路径
func ExecPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}

	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		path = strings.ReplaceAll(path, "\\", "/")
	}
	return path, err
}
