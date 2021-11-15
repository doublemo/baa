package session

import (
	"github.com/doublemo/baa/cores/log"

	logLocal "github.com/doublemo/baa/kits/agent/log"
)

// Logger 获取日志
func Logger() log.Logger {
	return logLocal.Logger()
}
