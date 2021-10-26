package sd

import "github.com/doublemo/baa/cores/sd"

// GetEndpointsByID 获取指定ID的节点
func GetEndpointsByID(values ...string) (map[string]sd.Endpoint, error) {
	eds, err := Endpoints()
	if err != nil {
		return nil, err
	}

	valuesMap := make(map[string]bool)
	for _, v := range values {
		valuesMap[v] = true
	}

	ret := make(map[string]sd.Endpoint)
	for _, endpoint := range eds {
		if valuesMap[endpoint.ID()] {
			ret[endpoint.ID()] = endpoint
		}
	}

	return ret, nil
}

// GetEndpointsByName 获取指定Name的节点
func GetEndpointsByName(values ...string) (map[string][]sd.Endpoint, error) {
	eds, err := Endpoints()
	if err != nil {
		return nil, err
	}

	valuesMap := make(map[string]bool)
	for _, v := range values {
		valuesMap[v] = true
	}

	ret := make(map[string][]sd.Endpoint)
	for _, endpoint := range eds {
		if valuesMap[endpoint.Name()] {
			if _, ok := ret[endpoint.Name()]; !ok {
				ret[endpoint.Name()] = make([]sd.Endpoint, 0)
			}

			ret[endpoint.Name()] = append(ret[endpoint.Name()], endpoint)
		}
	}
	return ret, nil
}
