package dao

import (
	"context"
	"errors"
	"math/rand"
	"strconv"

	"github.com/doublemo/baa/internal/sd"
)

const defaultAssignServerKey = "servers"

// AssignServers 为用户指定默认服务器
func AssignServers(ctx context.Context, userid uint64, serviceNames ...string) (map[string]string, error) {
	if rdb == nil {
		return nil, errors.New("rdb is nil")
	}

	namer := RDBNamer(defaultAssignServerKey, strconv.FormatUint(userid, 10))
	ret := rdb.HMGet(ctx, namer, serviceNames...)
	err := ret.Err()
	if err != nil {
		return nil, err
	}

	md := make(map[string]string)
	need := make([]string, 0)
	for k, v := range ret.Val() {
		if v == nil {
			need = append(need, serviceNames[k])
			continue
		}
		md[serviceNames[k]] = v.(string)
	}

	if len(need) < 1 {
		return md, nil
	}

	savedata := make(map[string]interface{})
	for _, name := range need {
		endpoints, err := sd.GetEndpointsByName(name)
		if err != nil {
			return nil, err
		}

		m, ok := endpoints[name]
		if !ok {
			continue
		}

		md[name] = m[rand.Intn(len(m))].ID()
		savedata[name] = md[name]
	}

	if len(savedata) > 0 {
		if err := rdb.HMSet(ctx, namer, savedata).Err(); err != nil {
			return nil, err
		}
	}

	return md, nil
}

// ReassignServers 重新分配服务器
func ReassignServers(ctx context.Context, userid uint64, serviceNames ...string) (map[string]string, error) {
	if rdb == nil {
		return nil, errors.New("rdb is nil")
	}

	namer := RDBNamer(defaultAssignServerKey, strconv.FormatUint(userid, 10))
	savedata := make(map[string]interface{})
	md := make(map[string]string)
	for _, name := range serviceNames {
		endpoints, err := sd.GetEndpointsByName(name)
		if err != nil {
			return nil, err
		}

		m, ok := endpoints[name]
		if !ok {
			continue
		}

		md[name] = m[rand.Intn(len(m))].ID()
		savedata[name] = md[name]
	}

	if len(savedata) > 0 {
		if err := rdb.HMSet(ctx, namer, savedata).Err(); err != nil {
			return nil, err
		}
	}

	return md, nil
}
