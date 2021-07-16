package sfu

import (
	"context"
	"fmt"
	"net"

	log "github.com/doublemo/baa/cores/log/level"
	coresnet "github.com/doublemo/baa/cores/net"
	"github.com/doublemo/baa/cores/os"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/rpcx"
	sfulog "github.com/pion/ion-sfu/pkg/logger"
	ionsfu "github.com/pion/ion-sfu/pkg/sfu"
)

// NewRPCXServerActor 提供RPC服务
func NewRPCXServerActor(config *conf.RPC, etcd *conf.Etcd, sfuconfig *Configuration) (*os.ProcessActor, error) {
	s, err := rpcx.NewRPCXServer(config.Salt, config.Key, config.ServiceSecurityKey)
	if err != nil {
		return nil, err
	}

	var c ionsfu.Config
	{
		c.SFU.Ballast = sfuconfig.Ballast
		c.SFU.WithStats = sfuconfig.WithStats
		c.WebRTC = sfuconfig.WebRTC
		c.Router = sfuconfig.Router
		c.Turn = sfuconfig.Turn
	}
	fmt.Println(c.WebRTC, c.Router, c.Turn)
	ionsfu.Logger = sfulog.New()
	serv := &sfuservice{
		server: s,
		sfu:    ionsfu.NewSFU(c),
	}

	addr, err := rpcxAddress(config.Addr)
	if err != nil {
		return nil, err
	}

	r, err := etcd.RPCXRegisterPlugin(addr)
	if err != nil {
		return nil, err
	}

	s.Plugins.Add(r)
	return &os.ProcessActor{
		Exec: func() error {
			Logger().Log("transport", "rpc", "on", config.Addr)
			s.RegisterName("sfu", serv, fmt.Sprintf("weight=%d&group=%s", config.Weight, config.Group))
			s.Serve(rpcx.Netname(), config.Addr)
			return nil
		},
		Interrupt: func(err error) {
			if err == nil {
				return
			}

			log.Error(Logger()).Log("transport", "rpc", "error", err)
		},

		Close: func() {
			Logger().Log("transport", "rpc", "on", "shutdown")
			if err := s.Shutdown(context.Background()); err != nil {
				log.Error(Logger()).Log("transport", "rpc", "error", err)
			}

			if err := s.UnregisterAll(); err != nil {
				log.Error(Logger()).Log("transport", "rpc", "error", err)
			}
		},
	}, nil
}

// rpcxAddress 获取服务地址
func rpcxAddress(addr string) (string, error) {
	ip, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", err
	}

	if ip == "" {
		// 尝试自动获取
		localIP, err := coresnet.LocalIP()
		if err != nil {
			return "", err
		}

		ip = localIP.String()
	}

	return rpcx.Netname() + "@" + net.JoinHostPort(ip, port), nil
}
