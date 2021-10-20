package server

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/cores/net"
	"github.com/doublemo/baa/cores/os"
	coressd "github.com/doublemo/baa/cores/sd"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/usrt"
	"github.com/doublemo/baa/kits/usrt/dao"
)

type Config struct {
	// MachineID 当前服务的唯一标识
	MachineID string `alias:"id" default:"snid01"`

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

	// Redis
	Redis conf.Redis `alias:"redis"`

	//Cacher cacher
	Cacher dao.CacherConfig `alias:"cacher"`

	// Nats
	Nats conf.Nats `alias:"nats"`
}

type USRT struct {
	// exitChan 退出信息
	exitChan chan struct{}

	// readyedChan 准备就绪信号
	readyedChan chan struct{}

	// configureOptions 配置文件
	configureOptions *ConfigureOptions

	//actors 服务进程管理
	actors *os.Process
}

func (s *USRT) Start() error {
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
	usrt.SetLogger(logger)

	// 服务发现
	endpoint := coressd.NewEndpoint(o.MachineID, usrt.ServiceName, o.RPC.Addr)
	endpoint.Set("group", o.RPC.Group)
	endpoint.Set("weight", strconv.Itoa(o.RPC.Weight))
	endpoint.Set("domain", o.Domain)
	endpoint.Set("ip", o.LocalIP)
	if err := sd.Init(o.Etcd, endpoint); err != nil {
		return err
	}

	// 检查机器码信息
	r, _ := regexp.Compile(`^[a-zA-Z]{1}\w+(\.\w+)+$`)
	if !r.MatchString(o.MachineID) {
		return errors.New("Invalid machineID:" + o.MachineID + ", eg:usrt1.cn.sc.cd")
	}

	if err := dao.Open(o.Redis, o.Cacher); err != nil {
		return fmt.Errorf("redis: %v", err)
	}

	// 路由
	usrt.InitRouter()
	o.Nats.Name = o.MachineID

	// 注册运行服务
	s.actors.Add(s.mustProcessActor(usrt.NewNatsProcessActor(o.Nats)), true)
	s.actors.Add(s.mustProcessActor(usrt.NewRPCServerActor(o.RPC)), true)
	s.actors.Add(s.mustProcessActor(usrt.NewServiceDiscoveryProcessActor()), true)
	return s.actors.Run()
}

func (s *USRT) Readyed() bool {
	return true
}

func (s *USRT) Shutdown() {
	s.actors.Stop()
}

func (s *USRT) Reload() {}

func (s *USRT) ServiceName() string {
	return ""
}

func (s *USRT) OtherCommand(cmd int) {

}

func (s *USRT) QuitCh() <-chan struct{} {
	return s.exitChan
}

func (s *USRT) mustProcessActor(actor *os.ProcessActor, err error) *os.ProcessActor {
	if err != nil {
		log.Error(logger).Log("error", err)
		panic(err)
	}

	return actor
}

// New 创建服务
func New(opts *ConfigureOptions) *USRT {
	return &USRT{
		exitChan:         make(chan struct{}),
		readyedChan:      make(chan struct{}),
		configureOptions: opts,
		actors:           os.NewProcess(),
	}
}