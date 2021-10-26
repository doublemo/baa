package dao

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/doublemo/baa/cores/cache/memcacher"
	"github.com/doublemo/baa/kits/usrt/proto/pb"
	"github.com/go-redis/redis/v8"
)

const (
	defaultUserStatusKey = "users"
	defaultMaxArgs       = 100
)

var (
	ErrArgsTooLong = errors.New("ErrArgsTooLong")
)

func namerStatusByUser(id uint64) string {
	mid := strconv.FormatUint(id, 10)
	return "status_user" + mid
}

// UpdateStatusByUser 更新用户状态
// return []uint64 所有未完成的设置的用户 error 错误
func UpdateStatusByUser(data ...*pb.USRT_User) ([]uint64, error) {
	if len(data) > defaultMaxArgs {
		return []uint64{}, ErrArgsTooLong
	}

	// 分组
	group := make(map[string]map[string]interface{})
	for _, item := range data {
		if item.ID < 1 {
			continue
		}

		id := strconv.FormatUint(item.ID, 10)
		if _, ok := group[id]; !ok {
			group[id] = make(map[string]interface{})
		}

		group[id][item.Type] = item.Value
		cacher.Delete(namerStatusByUser(item.ID))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer func() {
		cancel()
	}()

	values := make(map[string]*redis.BoolCmd)
	_, err := rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		for id, item := range group {
			values[id] = pipe.HMSet(ctx, RDBNamer(defaultUserStatusKey, id), item)
		}
		return nil
	})

	if err != nil {
		return []uint64{}, err
	}

	noCompleted := make([]uint64, 0)
	for id, ret := range values {
		if ret.Err() != nil || !ret.Val() {
			mid, _ := strconv.ParseUint(id, 10, 64)
			noCompleted = append(noCompleted, mid)
		}
	}

	return noCompleted, nil
}

// RemoveStatusByUser 删除用户状态
func RemoveStatusByUser(data ...*pb.USRT_User) error {
	if len(data) > defaultMaxArgs {
		return ErrArgsTooLong
	}

	// 分组
	group := make(map[string][]string)
	for _, item := range data {
		if item.ID < 1 {
			continue
		}

		id := strconv.FormatUint(item.ID, 10)
		if _, ok := group[id]; !ok {
			group[id] = make([]string, 0)
		}

		group[id] = append(group[id], item.Type)
		cacher.Delete(namerStatusByUser(item.ID))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer func() {
		cancel()
	}()

	_, err := rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		for id, item := range group {
			pipe.HDel(ctx, RDBNamer(defaultUserStatusKey, id), item...)
		}
		return nil
	})

	return err
}

func GetStatusByUser(data ...uint64) ([]*pb.USRT_User, error) {
	if len(data) > defaultMaxArgs {
		return nil, ErrArgsTooLong
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer func() {
		cancel()
	}()

	retx := make(map[uint64]*redis.StringStringMapCmd)
	_, err := rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, id := range data {
			retx[id] = pipe.HGetAll(ctx, RDBNamer(defaultUserStatusKey, strconv.FormatUint(id, 10)))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	values := make([]*pb.USRT_User, 0)
	for id, ret := range retx {
		if ret.Err() != nil {
			continue
		}

		for k, v := range ret.Val() {
			values = append(values, &pb.USRT_User{ID: id, Type: k, Value: v})
		}
	}

	return values, nil
}

func GetStatueCacheByUser(data ...uint64) ([]*pb.USRT_User, error) {
	values := make([]*pb.USRT_User, 0)
	noCache := make([]uint64, 0)
	for _, value := range data {
		cacheValues, ok := cacher.Get(namerStatusByUser(value))
		if !ok {
			noCache = append(noCache, value)
			continue
		}

		cacheValues2, ok := cacheValues.([]*pb.USRT_User)
		if !ok {
			noCache = append(noCache, value)
			continue
		}

		values = append(values, cacheValues2...)
	}

	if len(noCache) < 1 {
		return values, nil
	}

	retData, err := GetStatusByUser(noCache...)
	if err != nil {
		return nil, err
	}

	newrets := make([]*pb.USRT_User, len(values)+len(retData))
	if len(values) > 0 && len(retData) > 0 {
		copy(newrets[0:len(values)], values[0:])
		copy(newrets[len(values):], retData[0:])
	} else if len(values) > 0 {
		newrets = values
	} else {
		newrets = retData
	}

	// cache
	needCache := make(map[uint64][]*pb.USRT_User)
	for _, v := range retData {
		if _, ok := needCache[v.ID]; !ok {
			needCache[v.ID] = make([]*pb.USRT_User, 0)
		}
		needCache[v.ID] = append(needCache[v.ID], v)
	}

	if len(needCache) > 0 {
		for id, vv := range needCache {
			if len(vv) < 1 {
				continue
			}

			cacher.Set(namerStatusByUser(id), vv, memcacher.DefaultExpiration)
		}
	}
	return newrets, nil
}
