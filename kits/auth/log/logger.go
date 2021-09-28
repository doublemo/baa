package log

import (
	"os"

	"github.com/doublemo/baa/cores/log"
)

// logger 日志配置
var logger log.Logger

// SetLogger 设置日志
func SetLogger(l log.Logger) {
	logger = l
}

// Logger 获取日志
func Logger() log.Logger {

	if logger == nil {
		logger = log.NewLogfmtLogger(os.Stderr)
	}
	return logger
}
