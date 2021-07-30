// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

package conf

// RPC rpc参数
type RPC struct {
	// Name 服务名称
	Name string `alias:"name"`

	// Weight 调用权重
	Weight int `alias:"weight" default:"10"`

	// Group 分组
	Group string `alias:"group" default:"prod"`

	// Addr 监听地址
	Addr string `alias:"addr" default:":9092"`

	// StreamAddr 流服务监听地址
	StreamAddr string `alias:"streamaddr" default:":9093"`

	// StreamWaitNum 流服务等待队列最大容量
	StreamWaitNum int `alias:"streamwait" default:"1024"`

	// Salt 公钥
	// 当网络模式为tcp并且tls开启，这个支持将是pem的地址
	// 当网络模式为kcp时，这个将是kcp salt
	Salt string `alias:"salt" default:""`

	// Key 私钥
	// 当网络模式为tcp并且tls开启，这个支持将是key的地址
	// 当网络模式为kcp时，这个将是kcp key
	Key string `alias:"key" default:""`

	// ServiceSecurityKey 服务之通信认证
	ServiceSecurityKey string `alias:"sskey"`
}

// Clone RPC
func (o *RPC) Clone() *RPC {
	return &RPC{
		Name:               o.Name,
		Weight:             o.Weight,
		Group:              o.Group,
		Addr:               o.Addr,
		StreamAddr:         o.StreamAddr,
		StreamWaitNum:      o.StreamWaitNum,
		Salt:               o.Salt,
		Key:                o.Key,
		ServiceSecurityKey: o.ServiceSecurityKey,
	}
}

// RPCXClientPool RPXC Client pool
type RPCXClientPool struct {
	// PoolNumber 连接池数量
	PoolNumber int `alias:"pn" default:"10"`
}

// RPCXClient RPXC Client
type RPCXClient struct {
	// Group 分组
	Group string `alias:"group" default:"prod"`

	// Salt 公钥
	// 当网络模式为tcp并且tls开启，这个支持将是pem的地址
	// 当网络模式为kcp时，这个将是kcp salt
	Salt string `alias:"salt" default:""`

	// Key 私钥
	// 当网络模式为tcp并且tls开启，这个支持将是key的地址
	// 当网络模式为kcp时，这个将是kcp key
	Key string `alias:"key" default:""`

	// ServiceSecurityKey 服务之通信认证
	ServiceSecurityKey string `alias:"sskey"`

	// Pool 连接池
	Pool *RPCXClientPool `alias:"pool"`
}
