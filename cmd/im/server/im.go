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
	"github.com/doublemo/baa/internal/metrics"
	"github.com/doublemo/baa/internal/rpc"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/im"
	"github.com/doublemo/baa/kits/im/cache"
	"github.com/doublemo/baa/kits/im/dao"
	"github.com/doublemo/baa/kits/im/worker"
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

	Database conf.DBMySQLConfig `alias:"db"`

	// Router 路由
	Router im.RouterConfig `alias:"router"`

	// Cache 缓存
	Cache cache.CacherConfig `alias:"cache"`

	// Nats
	Nats conf.Nats `alias:"nats"`

	// workers 工人设置
	Worker worker.Config `alias:"worker"`

	// Metrics grpc metrics
	Metrics metrics.Config `alias:"metrics"`
}

type IM struct {
	// exitChan 退出信息
	exitChan chan struct{}

	// readyedChan 准备就绪信号
	readyedChan chan struct{}

	// configureOptions 配置文件
	configureOptions *ConfigureOptions

	//actors 服务进程管理
	actors *os.Process
}

func (s *IM) Start() error {
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
	im.SetLogger(logger)

	if o.Runmode == "dev" {
		o.Metrics.TurnOn = true
	}

	// 服务发现
	endpoint := coressd.NewEndpoint(o.MachineID, im.ServiceName, o.RPC.Addr)
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
		return errors.New("Invalid machineID:" + o.MachineID + ", eg:im1.cn.sc.cd")
	}

	if err := dao.Open(o.Database, o.Redis); err != nil {
		return fmt.Errorf("redis: %v", err)
	}

	// 缓存
	cache.Init(o.Cache)

	// 路由
	im.InitRouter(o.Router)
	o.Nats.Name = o.MachineID

	// 工人
	worker.Init(o.Worker)

	// 注册运行服务
	s.actors.Add(s.mustProcessActor(im.NewNatsProcessActor(o.Nats)), true)
	s.actors.Add(s.mustProcessActor(rpc.NewRPCServerActor(o.RPC, im.NewServer(), logger)), true)
	s.actors.Add(s.mustProcessActor(im.NewServiceDiscoveryProcessActor()), true)
	s.actors.Add(s.mustProcessActor(metrics.NewMetricsProcessActor(o.Metrics, logger)), true)
	return s.actors.Run()
}

func (s *IM) Readyed() bool {
	return true
}

func (s *IM) Shutdown() {
	s.actors.Stop()
}

func (s *IM) Reload() {}

func (s *IM) ServiceName() string {
	return ""
}

func (s *IM) OtherCommand(cmd int) {

}

func (s *IM) QuitCh() <-chan struct{} {
	return s.exitChan
}

func (s *IM) mustProcessActor(actor *os.ProcessActor, err error) *os.ProcessActor {
	if err != nil {
		log.Error(logger).Log("error", err)
		panic(err)
	}

	return actor
}

// New 创建服务
func New(opts *ConfigureOptions) *IM {
	return &IM{
		exitChan:         make(chan struct{}),
		readyedChan:      make(chan struct{}),
		configureOptions: opts,
		actors:           os.NewProcess(),
	}
}
