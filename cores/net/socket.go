// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/baa>

package net

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

// Socket 服务状态定义
const (
	// SocketStatusStoped socket 服务停止状态
	SocketStatusStoped int32 = iota

	// SocketStatusRunning socket 服务运行状态
	SocketStatusRunning
)

// Socket tcp 网络连接实现
type Socket struct {
	// done 服务运行完成信号
	done chan error

	// exit 退出信号
	exit chan struct{}

	// listen 监听
	listen *net.TCPListener

	// callback 连接回调
	callback func(net.Conn, chan struct{})

	// status 状态
	status int32

	wg sync.WaitGroup
}

// Serve 启动服务
func (s *Socket) Serve(addr string, readBufferSize, writeBufferSize int) error {
	s.done = make(chan error)
	s.exit = make(chan struct{})
	defer func() {
		close(s.done)
	}()

	if atomic.LoadInt32(&s.status) != SocketStatusStoped {
		return errors.New("ErrorServerNonStoped")
	}

	atomic.StoreInt32(&s.status, SocketStatusRunning)
	go s.serve(addr, readBufferSize, writeBufferSize)
	err := <-s.done
	close(s.exit)
	s.listen.Close()

	// waiting ...
	s.wg.Wait()
	atomic.StoreInt32(&s.status, SocketStatusStoped)
	return err
}

func (s *Socket) serve(addr string, readBufferSize, writeBufferSize int) {
	if err := s.listenTo(addr); err != nil {
		s.done <- err
		return
	}

	if s.callback == nil {
		s.done <- errors.New("ErrorCallBackIsNil")
		return
	}

	connChan := make(chan *net.TCPConn, 128)
	go func() {
		defer func() {
			close(connChan)
		}()

		err := s.accept(connChan)
		if atomic.LoadInt32(&s.status) == SocketStatusRunning {
			s.done <- err
		}
	}()

	for {
		select {
		case conn, ok := <-connChan:
			if !ok {
				return
			}

			conn.SetReadBuffer(readBufferSize)
			conn.SetWriteBuffer(writeBufferSize)
			go s.client(conn)

		case <-s.exit:
			return
		}
	}
}

func (s *Socket) accept(connChan chan *net.TCPConn) error {
	for {
		conn, err := s.listen.AcceptTCP()
		if err != nil {
			return err
		}

		connChan <- conn
	}
}

func (s *Socket) listenTo(addr string) (err error) {
	var resolveAddr *net.TCPAddr
	{
		resolveAddr, err = net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			return
		}
	}

	s.listen, err = net.ListenTCP("tcp", resolveAddr)
	return
}

func (s *Socket) client(conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
	}()

	s.wg.Add(1)
	s.callback(conn, s.exit)
}

// Status 状态
func (s *Socket) Status() int32 {
	return atomic.LoadInt32(&s.status)
}

// CallBack 设置回调方法
func (s *Socket) CallBack(f func(net.Conn, chan struct{})) {
	s.callback = f
}

// Shutdown 关闭
func (s *Socket) Shutdown() {
	if atomic.LoadInt32(&s.status) != SocketStatusRunning {
		return
	}
	atomic.StoreInt32(&s.status, SocketStatusStoped)
	s.done <- nil
}

// NewSocket 创建tcp socket
func NewSocket() *Socket {
	return &Socket{}
}
