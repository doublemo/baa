package server

import (
	"os"
	"time"

	"github.com/doublemo/baa/cores/log"
	kitlog "github.com/doublemo/baa/cores/log/level"
)

var logger log.Logger

func init() {
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.WithPrefix(logger, "[kit]", "sm")
	logger = log.With(logger, "ts", log.TimestampFormat(func() time.Time { return time.Now() }, time.RFC3339Nano))
}

// Logger 日志模式设置. dev 为开发模式
func Logger(runmode string) {
	if runmode == "dev" {
		logger = kitlog.NewFilter(logger, kitlog.AllowAll())
	} else {
		logger = kitlog.NewFilter(logger, kitlog.AllowError(), kitlog.AllowWarn())
	}
	logger = log.With(logger, "caller", log.DefaultCaller)
}
