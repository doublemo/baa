// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/baa>

package conf

// Http http配置
type Http struct {
	// Addr 监听地址
	Addr string `alias:"addr" default:":9090"`

	// ReadTimeout http服务读取超时
	ReadTimeout int `alias:"readtimeout" default:"10"`

	// WriteTimeout http服务写入超时
	WriteTimeout int `alias:"writetimeout" default:"10"`

	// MaxHeaderBytes  http内容大小限制
	MaxHeaderBytes int `alias:"maxheaderbytes" default:"1048576"`

	// SSL ssl 支持
	SSL bool `alias:"ssl" default:"false"`

	// Key 证书key
	Key string `alias:"key"`

	// SSLCert 证书
	Cert string `alias:"cert"`
}

// Clone Http
func (o *Http) Clone() *Http {
	return &Http{
		Addr:           o.Addr,
		ReadTimeout:    o.ReadTimeout,
		WriteTimeout:   o.WriteTimeout,
		MaxHeaderBytes: o.MaxHeaderBytes,
		SSL:            o.SSL,
		Key:            o.Key,
		Cert:           o.Cert,
	}
}
