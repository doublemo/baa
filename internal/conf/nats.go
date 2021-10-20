package conf

type (
	// Nats nats-io
	Nats struct {
		// Name 订阅客户端名称
		Name string `alias:"-"`

		// Urls 集群连接地址
		Urls []string `alias:"urls"`

		// MaxReconnects 最大重连次数
		MaxReconnects int `alias:"maxreconnects" default:"600"`

		// ReconnectWait 重连等待时间 秒为单位
		ReconnectWait int `alias:"reconnectwait" default:"1"`

		// ReconnectJitter 设置重连抖动时间 数组1 为非tls连接/毫秒 数组2 tls /秒
		ReconnectJitter []int `alias:"reconnectjitter"`

		// PingInterval ping 秒
		PingInterval int `alias:"pingInterval" default:"300"`

		// Authentication 安全验证
		Authentication NatsAuthentication `alias:"authentication"`

		// ChanSubscribeBuffer 通道订阅缓冲区大小
		ChanSubscribeBuffer int `alias:"chanSubscribeBuffer" default:"1"`
	}

	// NatsAuthentication 安装验证
	NatsAuthentication struct {
		// UserCreds 用户验证
		UserCreds string `alias:"usercreds"`

		// NK验证
		NkeyFile string `alias:"nkeyfile"`

		// TLSClientCert ssl证书
		TLSClientCert string `alias:"tlsclientcert"`

		// TLSClientkey ssl秘钥
		TLSClientKey string `alias:"tlsclientkey"`

		// TLSCACert 根证书
		TLSCACert string `alias:"tlscacert"`
	}
)
