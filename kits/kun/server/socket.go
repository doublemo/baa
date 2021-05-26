// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

package server

import (
	"net"

	"github.com/doublemo/baa/cores/log"
	kitlog "github.com/doublemo/baa/cores/log/level"
	coresnet "github.com/doublemo/baa/cores/net"
	"github.com/doublemo/baa/cores/os"
	"github.com/doublemo/baa/kits/kun/mem"
)

// NewServerSocket 创建socket
func NewServerSocket(o *mem.Parameters, callback func(net.Conn, chan struct{}), logger log.Logger) (*os.ProcessActor, error) {
	if o.Socket == nil {
		return nil, nil
	}

	var socket coresnet.Socket
	{
		socket.CallBack(callback)
	}

	return &os.ProcessActor{
		Exec: func() error {
			logger.Log("transport", "socket", "on", o.Socket.Addr)
			return socket.Serve(o.Socket.Addr, o.Socket.ReadBufferSize, o.Socket.WriteBufferSize)
		},
		Interrupt: func(err error) {
			if err == nil {
				return
			}

			kitlog.Error(logger).Log("transport", "socket", "error", err)
		},

		Close: func() {
			logger.Log("transport", "socket", "on", "shutdown")
			socket.Shutdown()
		},
	}, nil
}
