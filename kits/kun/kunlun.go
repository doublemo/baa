// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/baa>

package kun

import (
	"math/rand"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/net"
	"github.com/doublemo/baa/cores/os"
	"github.com/doublemo/baa/kits/kun/server"
)

type kun struct {
	// exitChan 退出信息
	exitChan chan struct{}

	// readyedChan 准备就绪信号
	readyedChan chan struct{}

	// configureOptions 配置文件
	configureOptions *ConfigureOptions

	//actors 服务进程管理
	actors *os.Process
}

func (s *kun) Start() error {
	defer func() {
		close(s.exitChan)
	}()

	rand.Seed(time.Now().UnixNano())

	// 读取一个配置文件副本
	o := s.configureOptions.Read()
	if o.LocalIP == "" {
		if m, err := net.LocalIP(); err == nil {
			o.LocalIP = m.String()
		}
	}

	// 设置日志
	Logger(o.Runmode)

	// 注册运行服务
	s.actors.Add(s.mustProcessActor(server.NewServerHttp(o, httpRouter(o), logger)), true)
	s.actors.Add(s.mustProcessActor(server.NewServerWebsocket(o, websocketRouter(o), logger)), true)
	s.actors.Add(s.mustProcessActor(server.NewServerSocket(o, socketRouter(o), logger)), true)
	return s.actors.Run()
}

func (s *kun) Readyed() bool {
	return true
}

func (s *kun) Shutdown() {}

func (s *kun) Reload() {}

func (s *kun) ServiceName() string {
	return ""
}

func (s *kun) OtherCommand(cmd int) {

}

func (s *kun) QuitCh() <-chan struct{} {
	return s.exitChan
}

func (s *kun) mustProcessActor(actor *os.ProcessActor, err error) *os.ProcessActor {
	if err != nil {
		log.Error(logger).Log("error", err)
		panic(err)
	}

	return actor
}

// New 创建服务
func New(opts *ConfigureOptions) *kun {
	return &kun{
		exitChan:         make(chan struct{}),
		readyedChan:      make(chan struct{}),
		configureOptions: opts,
		actors:           os.NewProcess(),
	}
}
