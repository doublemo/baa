// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/baa>

package mem

import "github.com/doublemo/baa/internal/conf"

// Parameters 参数
type Parameters struct {
	// MachineID 当前服务的唯一标识
	MachineID string `alias:"id" default:"kun01"`

	// Runmode 运行模式
	Runmode string `alias:"runmode" default:"pord"`

	// LocalIP 当前服务器IP地址
	LocalIP string `alias:"localip"`

	// Domain string 提供服务的域名
	Domain string `alias:"domain"`

	// Http http(s) 监听端口
	// 利用http实现信息GET/POST, webscoket 也会这个端口甚而上实现
	Http *conf.Http `alias:"http"`

	// Websocket 将支持WebSocket服务
	Websocket *conf.Webscoket `alias:"websocket"`

	// Socket 将支持tcp流服务
	Socket *conf.Scoket `alias:"socket"`
}
