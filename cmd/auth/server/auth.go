package server

import (
	"errors"
	"math/rand"
	"regexp"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/net"
	"github.com/doublemo/baa/cores/os"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/auth"
)

type Config struct {
	// MachineID 当前服务的唯一标识
	MachineID string `alias:"id" default:"kun01"`

	// Runmode 运行模式
	Runmode string `alias:"runmode" default:"pord"`

	// LocalIP 当前服务器IP地址
	LocalIP string `alias:"localip"`

	// Domain string 提供服务的域名
	Domain string `alias:"domain"`

	// RPC rpc
	RPC conf.RPC `alias:"rpc"`

	// Etcd etcd
	Etcd conf.Etcd `alias:"etcd"`

	// Nats
	Nats conf.Nats `alias:"nats"`

	// Router 路由
	Router auth.RouterConfig `alias:"router"`
}

type Authentication struct {
	// exitChan 退出信息
	exitChan chan struct{}

	// readyedChan 准备就绪信号
	readyedChan chan struct{}

	// configureOptions 配置文件
	configureOptions *ConfigureOptions

	//actors 服务进程管理
	actors *os.Process
}

func (s *Authentication) Start() error {
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
	auth.SetLogger(logger)

	// 服务发现
	if err := sd.Init(o.MachineID, o.Etcd, o.RPC); err != nil {
		return err
	}

	// 检查机器码信息
	r, _ := regexp.Compile(`^[a-zA-Z]{1}\w+(\.\w+)+$`)
	if !r.MatchString(o.MachineID) {
		return errors.New("Invalid machineID:" + o.MachineID + ", eg:auth1.cn.sc.cd")
	}

	o.Nats.Name = o.MachineID

	// 路由
	auth.InitRouter(o.Router)

	// 注册运行服务
	s.actors.Add(s.mustProcessActor(auth.NewNatsProcessActor(o.Nats)), true)
	s.actors.Add(s.mustProcessActor(auth.NewRPCServerActor(o.RPC)), true)
	s.actors.Add(s.mustProcessActor(auth.NewServiceDiscoveryProcessActor()), true)
	return s.actors.Run()
}

func (s *Authentication) Readyed() bool {
	return true
}

func (s *Authentication) Shutdown() {
	s.actors.Stop()
}

func (s *Authentication) Reload() {}

func (s *Authentication) ServiceName() string {
	return ""
}

func (s *Authentication) OtherCommand(cmd int) {

}

func (s *Authentication) QuitCh() <-chan struct{} {
	return s.exitChan
}

func (s *Authentication) mustProcessActor(actor *os.ProcessActor, err error) *os.ProcessActor {
	if err != nil {
		log.Error(logger).Log("error", err)
		panic(err)
	}

	return actor
}

// New 创建服务
func New(opts *ConfigureOptions) *Authentication {
	return &Authentication{
		exitChan:         make(chan struct{}),
		readyedChan:      make(chan struct{}),
		configureOptions: opts,
		actors:           os.NewProcess(),
	}
}
