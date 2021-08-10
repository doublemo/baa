package sd

import (
	"net/url"
)

type (
	// Endpoint 节点
	Endpoint interface {
		// ID 每个节点中唯一识别码
		ID() string

		// Name 服务名称
		Name() string

		// Addr 节点地址
		Addr() string

		// Set 设置节点附加内容
		Set(key, value string)

		// Get 获取节点附加内容
		Get(key string) string

		// Marshal 节点信息编码
		Marshal() string

		// Unmarshal 节点信息解码
		Unmarshal(string) error
	}

	// EndpointLocal 实现节点信息
	EndpointLocal struct {
		EId    string
		EName  string
		EAddr  string
		values url.Values
	}
)

func (endpoint *EndpointLocal) ID() string {
	return endpoint.EId
}

func (endpoint *EndpointLocal) Name() string {
	return endpoint.EName
}

func (endpoint *EndpointLocal) Addr() string {
	return endpoint.EAddr
}

func (endpoint *EndpointLocal) Get(key string) string {
	return endpoint.values.Get(key)
}

func (endpoint *EndpointLocal) Set(key, value string) {
	endpoint.values.Set(key, value)
}

func (endpoint *EndpointLocal) Marshal() string {
	endpoint.values.Set("id", endpoint.EId)
	endpoint.values.Set("name", endpoint.EName)
	endpoint.values.Set("addr", endpoint.EAddr)
	return endpoint.values.Encode()
}

func (endpoint *EndpointLocal) Unmarshal(data string) error {
	values, err := url.ParseQuery(data)
	if err != nil {
		return err
	}

	newValues := make(url.Values)
	for k, v := range values {

		if len(v) < 1 {
			continue
		}

		switch k {
		case "id":
			endpoint.EId = v[0]
		case "name":
			endpoint.EName = v[0]
		case "addr":
			endpoint.EAddr = v[0]
		default:
			newValues.Set(k, v[0])
		}
	}

	endpoint.values = newValues
	return nil
}

// NewEndpoint 创建节点
func NewEndpoint(id, name, addr string) *EndpointLocal {
	return &EndpointLocal{
		EId:    id,
		EName:  name,
		EAddr:  addr,
		values: make(url.Values),
	}
}
