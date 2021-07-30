package sd

import (
	"net/url"
)

const (
	FEndpointId     = "id"
	FEndpointName   = "name"
	FEndpointAddr   = "addr"
	FEndpointGroup  = "group"
	FEndpointWeight = "weight"
)

var endpointValidFields = map[string]bool{
	FEndpointId:     true,
	FEndpointName:   true,
	FEndpointAddr:   true,
	FEndpointGroup:  true,
	FEndpointWeight: true,
}

type (
	// Endpoint 节点
	Endpoint interface {
		// ID 每个节点中唯一识别码
		ID() string

		// Name 服务名称
		Name() string

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
		values url.Values
	}
)

func (endpoint *EndpointLocal) ID() string {
	return endpoint.values.Get(FEndpointId)
}

func (endpoint *EndpointLocal) Name() string {
	return endpoint.values.Get(FEndpointName)
}

func (endpoint *EndpointLocal) Get(key string) string {
	if !endpointValidFields[key] {
		return ""
	}

	return endpoint.values.Get(key)
}

func (endpoint *EndpointLocal) Set(key, value string) {
	if !endpointValidFields[key] {
		return
	}

	endpoint.values.Set(key, value)
}

func (endpoint *EndpointLocal) Marshal() string {
	return endpoint.values.Encode()
}

func (endpoint *EndpointLocal) Unmarshal(data string) error {
	values, err := url.ParseQuery(data)
	if err != nil {
		return err
	}

	newValues := make(url.Values)
	for k, v := range values {
		if !endpointValidFields[k] {
			continue
		}

		if len(v) < 1 {
			continue
		}

		newValues.Set(k, v[0])
	}

	endpoint.values = newValues
	return nil
}

func NewEndpoint(id, name string) *EndpointLocal {
	values := make(url.Values)
	values.Set(FEndpointId, id)
	values.Set(FEndpointName, name)
	return &EndpointLocal{
		values: values,
	}
}
