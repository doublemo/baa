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

	// Salt 公钥
	// 当网络模式为tcp并且tls开启，这个支持将是pem的地址
	Salt string `alias:"salt" default:""`

	// Key 私钥
	// 当网络模式为tcp并且tls开启，这个支持将是key的地址
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
		Salt:               o.Salt,
		Key:                o.Key,
		ServiceSecurityKey: o.ServiceSecurityKey,
	}
}

type RPCClient struct {
	// Name 服务名称
	Name string `alias:"name"`

	// Weight 调用权重
	Weight int `alias:"weight" default:"10"`

	// Group 分组
	Group string `alias:"group" default:"prod"`

	// Salt 公钥
	// 当网络模式为tcp并且tls开启，这个支持将是pem的地址
	Salt string `alias:"salt" default:""`

	// Key 私钥
	// 当网络模式为tcp并且tls开启，这个支持将是key的地址
	Key string `alias:"key" default:""`

	// ServiceSecurityKey 服务之通信认证
	ServiceSecurityKey string `alias:"sskey"`
}
