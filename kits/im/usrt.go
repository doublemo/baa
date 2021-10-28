package im

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	log "github.com/doublemo/baa/cores/log/level"
	grpcpool "github.com/doublemo/baa/cores/pool/grpc"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/nats"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/internal/rpc"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/im/cache"
	usrt "github.com/doublemo/baa/kits/usrt"
	grpcproto "github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

type usrtRouter struct {
	c     conf.RPCClient
	pool  *grpcpool.Pool
	mutex sync.RWMutex
}

func (r *usrtRouter) Serve(req *corespb.Request) (*corespb.Response, error) {
	var err error
	r.mutex.RLock()
	p := r.pool
	r.mutex.RUnlock()

	if p == nil {
		p, err = r.createPool()
		if err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := p.Get(ctx)
	if err != nil {
		return nil, err
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*5)
	defer func() {
		conn.Close()
		cancel2()
	}()

	client := corespb.NewServiceClient(conn.ClientConn)
	resp, err := client.Call(ctx2, req)
	return resp, err
}

func (r *usrtRouter) ServeHTTP(req *corespb.Request) (*corespb.Response, error) {
	return nil, errors.New("notsupported")
}

func (r *usrtRouter) createPool() (*grpcpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	fn := func(ctx context.Context) (*grpc.ClientConn, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		conn, err := rpc.NewConnect(r.c)
		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	p, err := grpcpool.NewWithContext(ctx, fn, r.c.Pool.Init, r.c.Pool.Capacity, time.Duration(r.c.Pool.IdleTimeout)*time.Minute, time.Duration(r.c.Pool.MaxLife)*time.Minute)
	if err != nil {
		return nil, err
	}

	r.mutex.Lock()
	if r.pool == nil {
		r.pool = p
	} else {
		p = r.pool
	}
	r.mutex.Unlock()
	return p, nil
}

func newUSRTRouter(c conf.RPCClient) *usrtRouter {
	return &usrtRouter{c: c}
}

func updateUserStatus(values ...*pb.USRT_User) ([]*pb.USRT_User, error) {
	if len(values) > 100 {
		return nil, errors.New("the value length cannot be greater then 100")
	}

	frame := pb.USRT_Status_Update{Values: values}
	b, _ := grpcproto.Marshal(&frame)
	resp, err := muxRouter.Handler(usrt.ServiceName, &corespb.Request{Command: command.USRTUpdateUserStatus.Int32(), Payload: b, Header: make(map[string]string)})
	if err != nil {
		return nil, err
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		resp := pb.USRT_Status_Reply{}
		if err := grpcproto.Unmarshal(payload.Content, &resp); err != nil {
			return nil, err
		}
		return resp.Values, nil

	case *corespb.Response_Error:
		return nil, errors.New(payload.Error.Message)
	}

	return nil, errors.New("usrt failed")
}

func deleteUserStatus(values ...*pb.USRT_User) {
	endpointer := sd.Endpointer()
	if endpointer == nil {
		return
	}

	endpoints, err := endpointer.Endpoints()
	if err != nil {
		log.Error(Logger()).Log("action", "kickedOut", "error", err)
		return
	}

	nc := nats.Conn()
	if nc == nil {
		return
	}

	frame := pb.USRT_Status_Update{Values: values}
	frameBytes, _ := grpcproto.Marshal(&frame)
	r := corespb.Request{
		Command: command.USRTDeleteUserStatus.Int32(),
		Payload: frameBytes,
		Header:  make(map[string]string),
	}

	r.Header["service"] = ServiceName
	r.Header["addr"] = sd.Endpoint().Addr()
	wBytes, _ := grpcproto.Marshal(&r)
	for _, endpoint := range endpoints {
		if endpoint.Name() != usrt.ServiceName {
			continue
		}
		nc.Publish(endpoint.ID(), wBytes)
	}
}

func getUserStatus(noCache bool, values ...uint64) ([]*pb.USRT_User, error) {
	if len(values) > 100 {
		return nil, errors.New("the value length cannot be greater then 100")
	}

	header := make(map[string]string)
	if noCache {
		header["no-cache"] = "true"
	}

	frame := pb.USRT_Status_Request{Values: values}
	b, _ := grpcproto.Marshal(&frame)
	resp, err := muxRouter.Handler(usrt.ServiceName, &corespb.Request{Command: command.USRTGetUserStatus.Int32(), Payload: b, Header: header})
	if err != nil {
		return nil, err
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		resp := pb.USRT_Status_Reply{}
		if err := grpcproto.Unmarshal(payload.Content, &resp); err != nil {
			return nil, err
		}
		return resp.Values, nil

	case *corespb.Response_Error:
		return nil, errors.New(payload.Error.Message)
	}
	return nil, errors.New("usrt failed")
}

func getCacheUserStatus(noCache bool, values ...uint64) (map[uint64]map[string]string, error) {
	data := make(map[uint64]map[string]string, 0)
	noCacheData := make([]uint64, 0)

	if !noCache {
		for _, value := range values {
			if m, ok := cache.Get(namerUserStatus(value)); ok && m != nil {
				if m0, ok := m.(map[string]string); ok {
					data[value] = m0
					continue
				}
			}
			noCacheData = append(noCacheData, value)
		}
	} else {
		noCacheData = values
	}

	if len(noCacheData) < 1 {
		return data, nil
	}

	retValues, err := getUserStatus(noCache, noCacheData...)
	if err != nil {
		return nil, err
	}

	newData := make(map[uint64]map[string]string, 0)
	for _, value := range retValues {
		if _, ok := newData[value.ID]; !ok {
			newData[value.ID] = make(map[string]string)
		}
		newData[value.ID][value.Type] = value.Value
	}

	for id, value := range newData {
		data[id] = value
		cache.Set(namerUserStatus(id), value, 0)
	}
	return data, nil
}

func namerUserStatus(id uint64) string {
	return "userstatus_" + strconv.FormatUint(id, 10)
}

func resetUserStatusCache(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.USRT_Status_Request
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	for _, id := range frame.Values {
		cache.Remove(namerUserStatus(id))
	}
	return nil, nil
}
