package conf

// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/baa>

// Scoket tcp参数
type Scoket struct {
	// Addr 监听地址
	Addr string `alias:"addr" default:":9091"`

	// ReadBufferSize 读取缓存大小 32767
	ReadBufferSize int `alias:"readbuffersize" default:"32767"`

	// WriteBufferSize 写入缓存大小 32767
	WriteBufferSize int `alias:"writebuffersize" default:"32767"`

	// ReadDeadline 读取超时
	ReadDeadline int `alias:"readdeadline" default:"310"`

	// WriteDeadline 写入超时
	WriteDeadline int `alias:"writedeadline"`

	// RPMLimit per connection rpm limit
	RPMLimit int `alias:"rpm" default:"200"`
}
