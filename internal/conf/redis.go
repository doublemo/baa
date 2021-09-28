package conf

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// Redis redis 配置
type Redis struct {
	// Addr redis集群地址
	Addr []string `alias:"addr"`

	Username string `alias:"username"`

	// Password 密码
	Password string `alias:"password"`

	DB int `alias:"db" default:"0"`

	// TLSClientCert ssl证书
	TLSClientCert string `alias:"tlsclientcert"`

	// TLSClientkey ssl秘钥
	TLSClientKey string `alias:"tlsclientkey"`

	// TLSServerName 如果您收到“x509：无法验证 xxx.xxx.xxx.xxx 的证书，因为它不包含任何 IP SAN”，请尝试设置
	TLSServerName string `alias:"tlsservername"`

	// Prefix 前缀
	Prefix string `alias:"prefix"`
}

// Connect 连接redis
func (r *Redis) Connect() (redis.UniversalClient, error) {
	o := &redis.UniversalOptions{Addrs: r.Addr, DB: r.DB}
	if r.Username != "" {
		o.Username = r.Username
	}

	if r.Password != "" {
		o.Password = r.Password
	}

	if r.TLSClientCert != "" && r.TLSClientKey != "" {
		cert, err := tls.LoadX509KeyPair(r.TLSClientCert, r.TLSClientKey)
		if err != nil {
			return nil, fmt.Errorf("redis: error loading client certificate: %v", err)
		}

		cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return nil, fmt.Errorf("redis: error parsing client certificate: %v", err)
		}

		o.TLSConfig = &tls.Config{
			MinVersion:   tls.VersionTLS12,
			ServerName:   "",
			Certificates: []tls.Certificate{cert},
		}

		if r.TLSServerName != "" {
			o.TLSConfig.ServerName = r.TLSServerName
		}
	}

	client := redis.NewUniversalClient(o)
	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	if pong != "PONG" {
		return nil, errors.New("redis ping error")
	}

	return client, nil
}
