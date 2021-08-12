package server

import (
	"math/rand"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/net"
	"github.com/doublemo/baa/cores/os"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/sfu"
)

type Config struct {
	// MachineID 当前服务的唯一标识
	MachineID string `alias:"id" default:"sfu01"`

	// Runmode 运行模式
	Runmode string `alias:"runmode" default:"pord"`

	// LocalIP 当前服务器IP地址
	LocalIP string `alias:"localip"`

	// Domain string 提供服务的域名
	Domain string `alias:"domain"`

	// Etcd etcd
	Etcd *conf.Etcd `alias:"etcd"`

	// RPC rpc
	RPC *conf.RPC `alias:"rpc"`

	// SFU ion-sfu
	SFU *sfu.Configuration `alias:"sfu"`
}

type SFU struct {
	// exitChan 退出信息
	exitChan chan struct{}

	// readyedChan 准备就绪信号
	readyedChan chan struct{}

	// configureOptions 配置文件
	configureOptions *ConfigureOptions

	//actors 服务进程管理
	actors *os.Process
}

func (s *SFU) Start() error {
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
	sfu.SetLogger(logger)
	sfu.InitRouter()

	// 服务发现
	if err := sd.Init(o.MachineID, o.Etcd.Clone(), o.RPC.Clone()); err != nil {
		return err
	}

	sfu.SetSFUServer(sfu.NewSFUServer(o.SFU))

	// 注册运行服务
	//s.actors.Add(s.mustProcessActor(sfu.NewRPCXServerActor(o.RPC.Clone(), o.Etcd.Clone(), o.SFU)), true)
	s.actors.Add(s.mustProcessActor(sfu.NewServerActor(o.RPC.Clone())), true)
	s.actors.Add(s.mustProcessActor(sfu.NewServiceDiscoveryProcessActor()), true)
	return s.actors.Run()
}

func (s *SFU) Readyed() bool {
	return true
}

func (s *SFU) Shutdown() {
	s.actors.Stop()
}

func (s *SFU) Reload() {}

func (s *SFU) ServiceName() string {
	return ""
}

func (s *SFU) OtherCommand(cmd int) {

}

func (s *SFU) QuitCh() <-chan struct{} {
	return s.exitChan
}

func (s *SFU) mustProcessActor(actor *os.ProcessActor, err error) *os.ProcessActor {
	if err != nil {
		log.Error(logger).Log("error", err)
		panic(err)
	}

	return actor
}

// New 创建服务
func New(opts *ConfigureOptions) *SFU {
	return &SFU{
		exitChan:         make(chan struct{}),
		readyedChan:      make(chan struct{}),
		configureOptions: opts,
		actors:           os.NewProcess(),
	}
}
