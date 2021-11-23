package dao

import (
	"context"
	"errors"
	"strconv"

	"github.com/go-redis/redis/v8"
)

const (
	defaultUsersStatusKey  = "users"
	defaultUserOnlineKey   = "online"
	defaultAssignServerKey = "servers"
)

type (
	// Users 用户数据
	Users struct {
		ID          uint64
		AgentServer string // 网关服务器
		Platform    string // 平台
		Token       string // 用户token
		OnlineAt    int64  // 上线时间
		IMServer    string // 聊天服务器
		IDServer    string // ID 分发服务器
	}
)

// GetPlatform 获取用户所有在线设备
func GetPlatform(ctx context.Context, userid uint64) ([]string, error) {
	if rdb == nil {
		return nil, errors.New("rdb is nil")
	}

	namer := RDBNamer(defaultUserOnlineKey, strconv.FormatUint(userid, 10))
	ret := rdb.SMembers(ctx, namer)
	return ret.Val(), ret.Err()
}

// GetUsers 获取所有的设备信息
func GetUsers(ctx context.Context, userid uint64) ([]*Users, error) {
	platform, err := GetPlatform(ctx, userid)
	if err != nil {
		return nil, err
	}

	idx := strconv.FormatUint(userid, 10)
	retx := make([]*redis.StringStringMapCmd, 0)
	_, err = rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, f := range platform {
			retx = append(retx, pipe.HGetAll(ctx, RDBNamer(defaultUsersStatusKey, idx, f)))
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	values := make([]*Users, 0)
	for _, ret := range retx {
		if ret.Err() != nil {
			continue
		}

		values = append(values, bindDataToUsers(ret.Val()))
	}
	return values, nil
}

func GetCachedUsers(ctx context.Context, userid uint64) ([]*Users, error) {
	data, ok := cacher.Get(makeUsersCacheID(userid))
	if ok {
		if m, ok := data.([]*Users); ok && m != nil {
			return m, nil
		}
	}
	return GetUsers(ctx, userid)
}

// GetMultiUsers 获取多用户
func GetMultiUsers(ctx context.Context, users ...uint64) (map[uint64][]*Users, error) {
	if rdb == nil {
		return nil, errors.New("rdb is nil")
	}

	platforms := make(map[uint64][]string)
	ret := make(map[uint64]*redis.StringSliceCmd)
	_, err := rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, id := range users {
			ret[id] = pipe.SMembers(ctx, RDBNamer(defaultUserOnlineKey, strconv.FormatUint(id, 10)))
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	for id, value := range ret {
		if value.Err() != nil {
			continue
		}

		platforms[id] = value.Val()
	}

	if len(platforms) < 1 {
		return nil, ErrRecordNotFound
	}

	retx := make(map[uint64][]*redis.StringStringMapCmd)
	_, err = rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		for id, values := range platforms {
			for _, value := range values {
				if _, ok := retx[id]; !ok {
					retx[id] = make([]*redis.StringStringMapCmd, 0)
				}

				retx[id] = append(retx[id], pipe.HGetAll(ctx, RDBNamer(defaultUsersStatusKey, strconv.FormatUint(id, 10), value)))
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	values := make(map[uint64][]*Users)
	for id, vv := range retx {
		for _, v := range vv {
			if v.Err() != nil {
				continue
			}

			if _, ok := values[id]; !ok {
				values[id] = make([]*Users, 0)
			}

			values[id] = append(values[id], bindDataToUsers(v.Val()))
		}
	}
	return values, nil
}

// GetCachedMultiUsers  从缓存中获取
func GetCachedMultiUsers(ctx context.Context, users ...uint64) (map[uint64][]*Users, error) {
	noCached := make([]uint64, 0)
	data := make(map[uint64][]*Users)
	for _, id := range users {
		c, ok := cacher.Get(makeUsersCacheID(id))
		if !ok {
			noCached = append(noCached, id)
			continue
		}

		user, ok := c.([]*Users)
		if !ok || user == nil {
			noCached = append(noCached, id)
			continue
		}

		data[id] = user
	}

	if len(noCached) < 1 {
		return data, nil
	}

	usersdata, err := GetMultiUsers(ctx, noCached...)
	if err != nil {
		return nil, err
	}

	for id, values := range usersdata {
		data[id] = values
		cacher.Set(makeUsersCacheID(id), values, 0)
	}

	return data, nil
}

// Online 用户上线
func Online(ctx context.Context, users *Users) error {
	if rdb == nil {
		return errors.New("rdb is nil")
	}

	idx := strconv.FormatUint(users.ID, 10)
	namerUsers := RDBNamer(defaultUsersStatusKey, idx, users.Platform)
	namerOnline := RDBNamer(defaultUserOnlineKey, idx)
	txf := func(tx *redis.Tx) error {
		ret := tx.SIsMember(ctx, namerOnline, users.Platform)
		if ret.Err() != nil {
			return ret.Err()
		}

		if ret.Val() {
			return ErrRecordIsFound
		}

		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.SAdd(ctx, namerOnline, users.Platform)
			pipe.HMSet(ctx, namerUsers, map[string]interface{}{
				"ID":          users.ID,
				"AgentServer": users.AgentServer,
				"Platform":    users.Platform,
				"Token":       users.Token,
				"OnlineAt":    users.OnlineAt,
				"IMServer":    users.IMServer,
				"IDServer":    users.IDServer,
			})

			return nil
		})
		return err
	}

	n := 0
loop:
	err := rdb.Watch(ctx, txf, namerOnline, namerUsers)
	if err == redis.TxFailedErr && n < 10 {
		n++
		goto loop
	}

	cacher.Delete(makeUsersCacheID(users.ID))
	return err
}

// Offline 用户下线
func Offline(ctx context.Context, userid uint64, platform string) error {
	if rdb == nil {
		return errors.New("rdb is nil")
	}

	idx := strconv.FormatUint(userid, 10)
	namerUsers := RDBNamer(defaultUsersStatusKey, idx, platform)
	namerOnline := RDBNamer(defaultUserOnlineKey, idx)

	_, err := rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.SRem(ctx, namerOnline, platform)
		pipe.Del(ctx, namerUsers)
		return nil
	})

	cacher.Delete(makeUsersCacheID(userid))
	return err
}

// UpdateUsersServer 更新用户服务地址
func UpdateUsersServer(ctx context.Context, userid uint64, server, addr string) error {
	if rdb == nil {
		return errors.New("rdb is nil")
	}

	idx := strconv.FormatUint(userid, 10)
	namer := RDBNamer(defaultAssignServerKey, idx)
	cacher.Delete(makeUsersCacheID(userid))
	return rdb.HSet(ctx, namer, server, addr).Err()
}

// GetUserServers 获取用户服务器
func GetUserServers(ctx context.Context, userid uint64) (map[string]string, error) {

	if rdb == nil {
		return nil, errors.New("rdb is nil")
	}

	idx := strconv.FormatUint(userid, 10)
	namer := RDBNamer(defaultAssignServerKey, idx)
	ret := rdb.HGetAll(ctx, namer)
	err := ret.Err()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	return ret.Val(), nil
}

// ClearUsersCachedByUserID 清除玩家状态信息缓存
func ClearUsersCachedByUserID(id uint64) {
	cacher.Delete(makeUsersCacheID(id))
}

func bindDataToUsers(data map[string]string) *Users {
	users := &Users{}
	if m, ok := data["ID"]; ok {
		users.ID, _ = strconv.ParseUint(m, 10, 64)
	}

	if m, ok := data["AgentServer"]; ok {
		users.AgentServer = m
	}

	if m, ok := data["Platform"]; ok {
		users.Platform = m
	}

	if m, ok := data["Token"]; ok {
		users.Token = m
	}

	if m, ok := data["OnlineAt"]; ok {
		users.OnlineAt, _ = strconv.ParseInt(m, 10, 64)
	}

	if m, ok := data["IMServer"]; ok {
		users.IMServer = m
	}

	if m, ok := data["IDServer"]; ok {
		users.IDServer = m
	}
	return users
}

func makeUsersCacheID(id uint64) string {
	return "cached_users_" + strconv.FormatUint(id, 10)
}
