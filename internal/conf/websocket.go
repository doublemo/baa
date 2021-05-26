// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/baa>

package conf

// Webscoket websocket配置
type Webscoket struct {
	// Addr 监听地址
	Addr string `alias:"addr" default:":9093"`

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

	// ReadBufferSize 读取缓存大小 32767
	ReadBufferSize int `alias:"readbuffersize" default:"32767"`

	// WriteBufferSize 写入缓存大小 32767
	WriteBufferSize int `alias:"writebuffersize" default:"32767"`

	// MaxMessageSize WebSocket每帧最在数据大小
	MaxMessageSize int64 `alias:"maxmessagesize" default:"1024"`

	// ReadDeadline 读取超时
	ReadDeadline int `alias:"readdeadline" default:"310"`

	// WriteDeadline 写入超时
	WriteDeadline int `alias:"writedeadline"`

	// RPMLimit per connection rpm limit
	RPMLimit int `alias:"rpm" default:"200"`
}

// Clone Webscoket
func (o *Webscoket) Clone() *Webscoket {
	return &Webscoket{
		Addr:            o.Addr,
		ReadTimeout:     o.ReadTimeout,
		WriteTimeout:    o.WriteTimeout,
		MaxHeaderBytes:  o.MaxHeaderBytes,
		SSL:             o.SSL,
		Key:             o.Key,
		Cert:            o.Cert,
		ReadBufferSize:  o.ReadBufferSize,
		WriteBufferSize: o.WriteBufferSize,
		MaxMessageSize:  o.MaxMessageSize,
		ReadDeadline:    o.ReadDeadline,
		WriteDeadline:   o.WriteDeadline,
		RPMLimit:        o.RPMLimit,
	}
}
