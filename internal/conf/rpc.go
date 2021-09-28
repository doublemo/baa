// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

package conf

type (
	// RPC rpc参数
	RPC struct {
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

	RPCClient struct {
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

		// pool 池
		Pool RPCPool
	}

	// RPCPool 池设置
	RPCPool struct {
		// Init 初始池中实例数量
		Init int `alias:"init" default:"1"`

		// Capacity 池最大容量
		Capacity int `alias:"capacity" default:"1"`

		// IdleTimeout 空闲超时/ 单位m
		IdleTimeout int `alias:"idleTimeout" default:"1"`

		// MaxLife 最大生命周期
		MaxLife int `alias:"maxlife" default:"1"`
	}
)
