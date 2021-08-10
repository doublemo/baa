package sd

import (
	"net/url"
	"testing"
)

func TestEndpoint(t *testing.T) {
	endpoint := &EndpointLocal{EId: "ddd", EName: "xxxxx", EAddr: "192.179.3.3:8980", values: make(url.Values)}
	endpoint.Set("weight", "10")
	data := endpoint.Marshal()
	t.Log(data)

	newendpoint := &EndpointLocal{}
	newendpoint.Unmarshal(data)
	t.Log(newendpoint)
}
