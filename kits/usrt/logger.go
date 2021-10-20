package usrt

import (
	"github.com/doublemo/baa/cores/log"
	logLocal "github.com/doublemo/baa/kits/usrt/log"
)

// SetLogger 设置日志
func SetLogger(logger log.Logger) {
	logLocal.SetLogger(logger)
}

// Logger 获取日志
func Logger() log.Logger {
	return logLocal.Logger()
}