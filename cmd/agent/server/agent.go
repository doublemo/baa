package server

import (
	"math/rand"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/net"
	"github.com/doublemo/baa/cores/os"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/kits/agent"
	"github.com/doublemo/baa/kits/agent/sd"
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

	// Http http(s) 监听端口
	// 利用http实现信息GET/POST, webscoket 也会这个端口甚而上实现
	Http *conf.Http `alias:"http"`

	// Websocket 将支持WebSocket服务
	Websocket *conf.Webscoket `alias:"websocket"`

	// Socket 将支持tcp流服务
	Socket *conf.Scoket `alias:"socket"`

	Router *agent.RouterConfig `alias:"router"`

	// RPC rpc
	RPC *conf.RPC `alias:"rpc"`

	// Etcd etcd
	Etcd *conf.Etcd `alias:"etcd"`
}

type Agent struct {
	// exitChan 退出信息
	exitChan chan struct{}

	// readyedChan 准备就绪信号
	readyedChan chan struct{}

	// configureOptions 配置文件
	configureOptions *ConfigureOptions

	//actors 服务进程管理
	actors *os.Process
}

func (s *Agent) Start() error {
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
	agent.SetLogger(logger)

	// 服务发现
	if err := sd.Init(o.MachineID, o.Etcd.Clone(), o.RPC.Clone()); err != nil {
		return err
	}

	// 路由
	agent.InitRouter(o.Router)

	// 初始数据

	// 注册运行服务

	//s.actors.Add(s.mustProcessActor(agent.NewSFUActor(o.Router.Sfu, o.Router.Etcd)), true)
	s.actors.Add(s.mustProcessActor(agent.NewSocketProcessActor(o.Socket.Clone())), true)
	s.actors.Add(s.mustProcessActor(agent.NewWebsocketProcessActor(o.Websocket.Clone())), true)
	s.actors.Add(s.mustProcessActor(agent.NewServiceDiscoveryProcessActor()), true)
	return s.actors.Run()
}

func (s *Agent) Readyed() bool {
	return true
}

func (s *Agent) Shutdown() {
	s.actors.Stop()
}

func (s *Agent) Reload() {}

func (s *Agent) ServiceName() string {
	return ""
}

func (s *Agent) OtherCommand(cmd int) {

}

func (s *Agent) QuitCh() <-chan struct{} {
	return s.exitChan
}

func (s *Agent) mustProcessActor(actor *os.ProcessActor, err error) *os.ProcessActor {
	if err != nil {
		log.Error(logger).Log("error", err)
		panic(err)
	}

	return actor
}

// New 创建服务
func New(opts *ConfigureOptions) *Agent {
	return &Agent{
		exitChan:         make(chan struct{}),
		readyedChan:      make(chan struct{}),
		configureOptions: opts,
		actors:           os.NewProcess(),
	}
}
