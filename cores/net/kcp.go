// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/baa>

package net

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"

	kcp "github.com/xtaci/kcp-go"
)

// KCP 服务状态定义
const (
	// SocketStatusStoped socket 服务停止状态
	KCPStatusStoped int32 = iota

	// SocketStatusRunning socket 服务运行状态
	KCPStatusRunning
)

// KCP go版本的kcp支持
type KCP struct {
	// done 服务运行完成信号
	done chan error

	// exit 退出信号
	exit chan struct{}

	// listen 监听
	listen *kcp.Listener

	// callback 连接回调
	callback func(net.Conn, chan struct{})

	// status 状态
	status int32

	//config  配置信息信息
	config *KCPConfig

	wg sync.WaitGroup
}

// Serve 启动服务
func (s *KCP) Serve(c *KCPConfig) error {
	s.config = c
	s.done = make(chan error)
	s.exit = make(chan struct{})
	defer func() {
		close(s.done)
	}()

	if atomic.LoadInt32(&s.status) != SocketStatusStoped {
		return errors.New("ErrorServerNonStoped")
	}

	atomic.StoreInt32(&s.status, SocketStatusRunning)
	go s.serve()
	err := <-s.done
	close(s.exit)
	s.listen.Close()

	// waiting ...
	s.wg.Wait()
	atomic.StoreInt32(&s.status, SocketStatusStoped)
	return err
}

func (s *KCP) serve() {
	if err := s.listenTo(); err != nil {
		s.done <- err
		return
	}

	if s.callback == nil {
		s.done <- errors.New("ErrorCallBackIsNil")
		return
	}

	connChan := make(chan *kcp.UDPSession, 128)
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

			conn.SetStreamMode(true)
			conn.SetWriteDelay(false)
			conn.SetNoDelay(s.config.Delay())
			conn.SetMtu(s.config.MTU)
			conn.SetWindowSize(s.config.SndWnd, s.config.RcvWnd)
			conn.SetACKNoDelay(s.config.AckNodelay)
			go s.client(conn)

		case <-s.exit:
			return
		}
	}
}

func (s *KCP) accept(connChan chan *kcp.UDPSession) error {
	for {
		conn, err := s.listen.AcceptKCP()
		if err != nil {
			return err
		}

		connChan <- conn
	}
}

func (s *KCP) listenTo() (err error) {
	block, err := s.config.BlockCrypt()
	if err != nil {
		return err
	}

	s.listen, err = kcp.ListenWithOptions(s.config.Addr, block, s.config.DataShard, s.config.ParityShard)
	if err != nil {
		return
	}

	if err := s.listen.SetDSCP(s.config.DSCP); err != nil {
		return err
	}

	if err := s.listen.SetReadBuffer(s.config.SockBuf); err != nil {
		return err
	}

	if err := s.listen.SetWriteBuffer(s.config.SockBuf); err != nil {
		return err
	}

	return
}

func (s *KCP) client(conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
	}()

	s.wg.Add(1)
	s.callback(conn, s.exit)
}

// Status 状态
func (s *KCP) Status() int32 {
	return atomic.LoadInt32(&s.status)
}

// CallBack 设置回调方法
func (s *KCP) CallBack(f func(net.Conn, chan struct{})) {
	s.callback = f
}

// Shutdown 关闭
func (s *KCP) Shutdown() {
	if atomic.LoadInt32(&s.status) != SocketStatusRunning {
		return
	}
	atomic.StoreInt32(&s.status, SocketStatusStoped)
	s.done <- nil
}

// NewKCP 创建kcp
func NewKCP() *KCP {
	return &KCP{}
}
