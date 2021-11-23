package server

import (
	"errors"
	"math/rand"
	"regexp"
	"strconv"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/net"
	"github.com/doublemo/baa/cores/os"
	coressd "github.com/doublemo/baa/cores/sd"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/metrics"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/agent"
	"github.com/doublemo/baa/kits/agent/webrtc"
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
	Http conf.Http `alias:"http"`

	// Websocket 将支持WebSocket服务
	Websocket conf.Webscoket `alias:"websocket"`

	// Socket 将支持tcp流服务
	Socket conf.Scoket `alias:"socket"`

	Router agent.RouterConfig `alias:"router"`

	// RPC rpc
	RPC conf.RPC `alias:"rpc"`

	// Etcd etcd
	Etcd conf.Etcd `alias:"etcd"`

	//Webrtc  webrtc config
	Webrtc webrtc.WebRTCConfig `alias:"webrtc"`

	// Nats
	Nats conf.Nats `alias:"nats"`

	// Metrics grpc metrics
	Metrics metrics.Config `alias:"metrics"`
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

	if o.Runmode == "dev" {
		o.Metrics.TurnOn = true
	}

	// 服务发现
	endpoint := coressd.NewEndpoint(o.MachineID, agent.ServiceName, o.RPC.Addr)
	endpoint.Set("group", o.RPC.Group)
	endpoint.Set("weight", strconv.Itoa(o.RPC.Weight))
	endpoint.Set("domain", o.Domain)
	endpoint.Set("ip", o.LocalIP)
	endpoint.Set("socket", o.Socket.Addr)
	endpoint.Set("http", o.Http.Addr)
	endpoint.Set("websocket", o.Websocket.Addr)
	if err := sd.Init(o.Etcd, endpoint); err != nil {
		return err
	}

	// 路由
	agent.InitRouter(o.Router)

	// webrtc
	if err := webrtc.Init(o.Webrtc); err != nil {
		return err
	}

	// 检查机器码信息
	r, _ := regexp.Compile(`^[a-zA-Z]{1}\w+(\.\w+)+$`)
	if !r.MatchString(o.MachineID) {
		return errors.New("Invalid machineID:" + o.MachineID + ", eg:agent1.cn.sc.cd")
	}

	o.Nats.Name = o.MachineID

	// 注册运行服务
	s.actors.Add(s.mustProcessActor(agent.NewNatsProcessActor(o.Nats)), true)
	s.actors.Add(s.mustProcessActor(agent.NewSocketProcessActor(o.Socket)), true)
	s.actors.Add(s.mustProcessActor(agent.NewWebsocketProcessActor(o.Websocket)), true)
	s.actors.Add(s.mustProcessActor(agent.NewHttpProcessActor(o.Http, o.Router)), true)
	s.actors.Add(s.mustProcessActor(agent.NewRPCServerActor(o.RPC)), true)
	s.actors.Add(s.mustProcessActor(agent.NewServiceDiscoveryProcessActor()), true)
	s.actors.Add(s.mustProcessActor(metrics.NewMetricsProcessActor(o.Metrics, logger)), true)
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
