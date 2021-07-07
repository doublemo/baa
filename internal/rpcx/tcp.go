// +build !kcp

package rpcx

import (
	"context"
	"crypto/tls"
	"errors"

	"github.com/doublemo/baa/internal/conf"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
)

// Use 设置中间件信息
func Use(xclient client.XClient, clientOption conf.RPCXClient) {
	if len(clientOption.ServiceSecurityKey) > 0 {
		xclient.Auth("bearer " + clientOption.ServiceSecurityKey)
	}
}

// Netname rpcx 网络名称
func Netname() string {
	return "tcp"
}

// NewRPCXServer 创建RPCX server
func NewRPCXServer(salt, key, sskey string) (*server.Server, error) {
	var s *server.Server
	if len(key) > 0 && len(salt) > 0 {
		cert, err := tls.LoadX509KeyPair(salt, key)
		if err != nil {
			return nil, err
		}

		config := &tls.Config{Certificates: []tls.Certificate{cert}}
		s = server.NewServer(server.WithTLSConfig(config))
	} else {
		s = server.NewServer()
	}

	if len(sskey) > 0 {
		s.AuthFunc = func(ctx context.Context, req *protocol.Message, token string) error {
			if token == "bearer "+sskey {
				return nil
			}

			return errors.New("InvalidSecurityToken")
		}
	}

	return s, nil
}

// Option 请求设置
func Option(clientOption conf.RPCXClient) client.Option {
	option := client.DefaultOption
	option.Retries = 10
	option.SerializeType = protocol.ProtoBuffer
	option.Group = "prod"

	if len(clientOption.Key) > 0 && len(clientOption.Salt) > 0 {
		option.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	return option
}
