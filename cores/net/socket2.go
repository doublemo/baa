package net

import (
	"net"
	"sync"
)

// Socket2 tcp 网络连接实现
type Socket2 struct {
	listen    *net.TCPListener
	onConnect func(net.Conn)
	onClose   func()
	stopOnce  sync.Once
	mutexRW   sync.RWMutex
	exit      chan struct{}
}

// OnConnect 设置回调方法
func (s *Socket2) OnConnect(f func(net.Conn)) {
	s.mutexRW.Lock()
	s.onConnect = f
	s.mutexRW.Unlock()
}

// OnClose 关闭时调用
func (s *Socket2) OnClose(f func()) {
	s.mutexRW.Lock()
	s.onClose = f
	s.mutexRW.Unlock()
}

// Serve 启动服务
func (s *Socket2) Serve(addr string, readBufferSize, writeBufferSize int) error {
	defer func() {
		s.mutexRW.RLock()
		onClose := s.onClose
		s.mutexRW.RUnlock()

		if onClose != nil {
			onClose()
		}
	}()
	if err := s.listenTo(addr); err != nil {
		return err
	}

	connChan := make(chan *net.TCPConn, 128)
	done := make(chan error)
	go func(c chan *net.TCPConn, ok chan error) {
		defer func() {
			close(c)
		}()

		ok <- s.accept(c)
	}(connChan, done)

	for {
		select {
		case conn, ok := <-connChan:
			if !ok {
				return nil
			}

			conn.SetReadBuffer(readBufferSize)
			conn.SetWriteBuffer(writeBufferSize)
			if s.onConnect != nil {
				s.onConnect(conn)
			}

		case err := <-done:
			return err

		case <-s.exit:
			return nil
		}
	}
}

func (s *Socket2) accept(connChan chan *net.TCPConn) error {
	for {
		s.mutexRW.RLock()
		listen := s.listen
		s.mutexRW.RUnlock()

		conn, err := listen.AcceptTCP()
		if err != nil {
			return err
		}

		select {
		case connChan <- conn:
		case <-s.exit:
			return nil
		}
	}
}

func (s *Socket2) listenTo(addr string) (err error) {
	var resolveAddr *net.TCPAddr
	{
		resolveAddr, err = net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			return
		}
	}

	s.mutexRW.Lock()
	s.listen, err = net.ListenTCP("tcp", resolveAddr)
	s.mutexRW.Unlock()
	return
}

// Shutdown 关闭
func (s *Socket2) Shutdown() {
	s.stopOnce.Do(func() {
		close(s.exit)
		s.mutexRW.RLock()
		s.listen.Close()
		s.mutexRW.RUnlock()
	})
}

// NewSocket2 创建tcp socket
func NewSocket2() *Socket2 {
	return &Socket2{
		exit: make(chan struct{}),
	}
}
