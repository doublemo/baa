// +build kcp

package rpcx

import (
	"context"
	"crypto/sha1"
	"errors"

	"github.com/doublemo/baa/internal/conf"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
)

// Netname rpcx 网络名称
func Netname() string {
	return "kcp"
}

// Use 设置中间件信息
func Use(xclient client.XClient, clientOption conf.RPCXClient) {
	if len(clientOption.ServiceSecurityKey) > 0 {
		xclient.Auth("bearer " + clientOption.ServiceSecurityKey)
	}

	cs := &UDPSession{}
	pc := client.NewPluginContainer()
	pc.Add(cs)
	xclient.SetPlugins(pc)
}

// NewRPCXServer 创建RPCX server
func NewRPCXServer(salt, key, sskey string) (*server.Server, error) {
	pass := pbkdf2.Key([]byte(key), []byte(salt), 4096, 32, sha1.New)
	bc, err := kcp.NewAESBlockCrypt(pass)
	if err != nil {
		return nil, err
	}

	s := server.NewServer(server.WithBlockCrypt(bc))
	cs := &UDPSession{}
	s.Plugins.Add(cs)

	if len(sskey) > 0 {
		s.AuthFunc = func(ctx context.Context, req *protocol.Message, token string) error {
			if token == "bearer "+sskey {
				return nil
			}

			return errors.New("InvalidSecurityToken")
		}
	}

	return s
}

// Option 请求设置
func Option(clientOption conf.RPCXClient) client.Option {
	option := client.DefaultOption
	option.Retries = 10
	option.SerializeType = protocol.ProtoBuffer

	if len(clientOption.Key) > 0 && len(clientOption.Salt) > 0 {
		pass := pbkdf2.Key([]byte(clientOption.Key), []byte(clientOption.Salt), 4096, 32, sha1.New)
		bc, _ := kcp.NewAESBlockCrypt(pass)
		option.Block = bc
	}
	return option
}
