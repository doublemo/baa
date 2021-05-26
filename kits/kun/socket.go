// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/baa>

package kun

import (
	"fmt"
	"net"
	"runtime"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/kits/kun/mem"
)

func socketRouter(o *mem.Parameters) func(net.Conn, chan struct{}) {
	return func(conn net.Conn, exit chan struct{}) {
		defer func() {

		}()

		socketLoop(o, exit)
	}
}

func socketLoop(o *mem.Parameters, exit chan struct{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(logger).Log("panic", fmt.Sprint(r))
			i := 0
			funcName, file, line, ok := runtime.Caller(i)
			for ok {
				log.Error(logger).Log("panic", fmt.Sprintf("frame %v:[func:%v,file:%v,line:%v]", i, runtime.FuncForPC(funcName).Name(), file, line))
				i++
				funcName, file, line, ok = runtime.Caller(i)
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer func() {
		ticker.Stop()
	}()

	//createAt := time.Now()
	packetCounter := 0
	rpm1Min := 0
	rpmLimit := o.Socket.RPMLimit
	for {
		select {
		case <-ticker.C:
			rpm1Min++
			if rpm1Min >= 60 {
				if packetCounter > rpmLimit {
					return
				}

				rpm1Min = 0
				packetCounter = 0
			}

		case <-exit:
			return
		}
	}
}
