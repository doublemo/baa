package auth

import (
	"context"
	"errors"
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
