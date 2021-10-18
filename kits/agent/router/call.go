package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/doublemo/baa/cores/crypto/id"
	coreslog "github.com/doublemo/baa/cores/log"
	log "github.com/doublemo/baa/cores/log/level"
	grpcpool "github.com/doublemo/baa/cores/pool/grpc"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/rpc"
	"github.com/doublemo/baa/kits/agent/errcode"
	"github.com/doublemo/baa/kits/agent/proto"
	"github.com/doublemo/baa/kits/agent/session"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

const (
	// defaultMaxQueryLength 默认http query最大长度
	defaultMaxQueryLength int = 1024

	// defaultMaxBytesReader 限制最大body大小
	defaultMaxBytesReader int64 = 1 << 20
)

type (
	// Call 处理路由Call方法
	Call struct {
		c              conf.RPCClient
		pool           *grpcpool.Pool
		onDestroy      atomic.Value
		onBeforeCall   atomic.Value
		onAfterCall    atomic.Value
		commandSecret  []byte
		maxQureyLength int
		maxBytesReader int64
		mutex          sync.RWMutex
		logger         coreslog.Logger
	}

	CallOptions func(c *Call)
)

// Serve 服务处理
func (r *Call) Serve(peer session.Peer, req coresproto.Request) (coresproto.Response, error) {
	request := &corespb.Request{
		Header:  map[string]string{"PeerId": peer.ID(), "seqno": strconv.FormatUint(uint64(req.SID()), 10), "Content-Type": "stream"},
		Command: req.SubCommand().Int32(),
		Payload: req.Body(),
	}

	if handler, ok := r.onBeforeCall.Load().(func(session.Peer, coresproto.Request, *corespb.Request) error); ok && handler != nil {
		if err := handler(peer, req, request); err != nil {
			return nil, err
		}
	}

	resp, err := r.Call(request)
	if err != nil {
		return proto.NewResponseBytes(req.Command(), errcode.Bad(&corespb.Response{Command: req.SubCommand().Int32()}, errcode.ErrInternalServer, grpc.ErrorDesc(err))), nil
	}

	fmt.Println("x99->", r.onAfterCall.Load())
	if handler, ok := r.onAfterCall.Load().(func(session.Peer, *corespb.Response) error); ok && handler != nil {
		fmt.Println("x99->222", handler)
		if err := handler(peer, resp); err != nil {
			return nil, err
		}
	}

	w := proto.NewResponseBytes(req.Command(), resp)
	return w, nil
}

func (r *Call) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	cmd := vars["command"]
	if cmd == "" {
		http.Error(rw, "feature not supported", http.StatusBadRequest)
		return
	}

	command, err := id.Decrypt(cmd, []byte(r.commandSecret))
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	request, err := r.buildCoresPBRequest(rw, req)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	request.Command = int32(command)
	request.Header["Content-Type"] = "json"
	request.Header["X-Session-Token"] = req.Header.Get("X-Session-Token")
	if handler, ok := r.onBeforeCall.Load().(func(session.Peer, coresproto.Request, *corespb.Request) error); ok && handler != nil {
		if err := handler(nil, nil, request); err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
	}

	resp, err := r.Call(request)
	if err != nil {
		log.Error(r.logger).Log("action", "call", "error", err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if handler, ok := r.onAfterCall.Load().(func(session.Peer, *corespb.Response) error); ok && handler != nil {
		if err := handler(nil, resp); err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
	}

	rw.Header().Set("Content-Type", "application/json;charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	switch rwPaylod := resp.Payload.(type) {
	case *corespb.Response_Error:
		ret, _ := json.Marshal(rwPaylod.Error)
		io.WriteString(rw, string(ret))

	case *corespb.Response_Content:
		io.WriteString(rw, string(rwPaylod.Content))
	}
}

// Call 调用
func (r *Call) Call(req *corespb.Request) (*corespb.Response, error) {
	p, err := r.createPool()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := p.Get(ctx)
	if err != nil {
		return nil, err
	}

	client := corespb.NewServiceClient(conn.ClientConn)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		conn.Close()
		cancel2()
	}()

	return client.Call(ctx2, req)
}

func (r *Call) createPool() (*grpcpool.Pool, error) {
	r.mutex.RLock()
	if r.pool != nil {
		r.mutex.RUnlock()
		return r.pool, nil
	}
	r.mutex.RUnlock()

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

// Destroy 清理
func (r *Call) Destroy(peer session.Peer) {
	if handler, ok := r.onDestroy.Load().(func(session.Peer)); ok && handler != nil {
		handler(peer)
	}
}

// OnDestroy 存储当路由清理时响应函数
func (r *Call) OnDestroy(f func(peer session.Peer)) {
	if f == nil {
		return
	}

	r.onDestroy.Store(f)
}

// OnBeforeCall Hook
func (r *Call) OnBeforeCall(f func(session.Peer, coresproto.Request, *corespb.Request) error) {
	if f == nil {
		return
	}

	r.onBeforeCall.Store(f)
}

// OnAfterCall Hook
func (r *Call) OnAfterCall(f func(session.Peer, *corespb.Response) error) {
	if f == nil {
		return
	}

	r.onAfterCall.Store(f)
}

func (r *Call) resolveContentType(req *http.Request) string {
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		return "text/html"
	}
	return strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
}

func (r *Call) buildCoresPBRequest(rw http.ResponseWriter, req *http.Request) (*corespb.Request, error) {
	switch strings.ToUpper(req.Method) {
	case "GET":
		return r.buildCoresPBMethodGet(req)

	case "POST":
		return r.buildCoresPBMethodPost(rw, req)
	}

	return nil, errors.New("invalid method")
}

func (r *Call) buildCoresPBMethodGet(req *http.Request) (*corespb.Request, error) {
	urlValues := req.URL.Query()
	request := &corespb.Request{
		Header:  make(map[string]string),
		Payload: make([]byte, 0),
	}

	if len(req.URL.RawQuery) > r.maxQureyLength {
		return nil, errors.New("URL query too long")
	}

	values := make(map[string]string)
	for k, v := range urlValues {
		values[k] = strings.Join(v, ";")
	}

	request.Header = values
	request.Payload = make([]byte, 0)
	return request, nil
}

func (r *Call) buildCoresPBMethodPost(rw http.ResponseWriter, req *http.Request) (*corespb.Request, error) {
	urlValues := req.URL.Query()
	request := &corespb.Request{
		Header:  make(map[string]string),
		Payload: make([]byte, 0),
	}

	if len(req.URL.RawQuery) > r.maxQureyLength {
		return nil, errors.New("URL query too long")
	}

	values := make(map[string]string)
	for k, v := range urlValues {
		values[k] = strings.Join(v, ";")
	}

	request.Header = values
	req.Body = http.MaxBytesReader(rw, req.Body, r.maxBytesReader)
	defer req.Body.Close()

	switch r.resolveContentType(req) {
	case "application/json", "application/x-protobuf":
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		request.Payload = body

	case "application/x-www-form-urlencoded":
		if err := req.ParseForm(); err != nil {
			return nil, err
		}
		formValues := make(map[string]string)
		for k, v := range req.Form {
			formValues[k] = strings.Join(v, ";")
		}

		body, err := json.Marshal(values)
		if err != nil {
			return nil, err
		}
		request.Payload = body

	default:
		return nil, errors.New("mime not supported")
	}

	return request, nil
}

// NewCall 创建Call路由
func NewCall(config conf.RPCClient, logger coreslog.Logger, opts ...CallOptions) *Call {
	c := &Call{
		c:              config,
		commandSecret:  make([]byte, 0),
		maxQureyLength: defaultMaxQueryLength,
		maxBytesReader: defaultMaxBytesReader,
		logger:         logger,
	}

	for _, o := range opts {
		o(c)
	}
	return c
}

// CommandSecretCallOptions 设置Command解决key
func CommandSecretCallOptions(secret string) CallOptions {
	return func(r *Call) {
		r.commandSecret = []byte(secret)
	}
}

// MaxQureyLengthCallOptions 设置最大http query 长度
func MaxQureyLengthCallOptions(num int) CallOptions {
	return func(r *Call) {
		r.maxQureyLength = num
	}
}

// MaxBytesReaderCallOptions 设置最大body size
func MaxBytesReaderCallOptions(num int64) CallOptions {
	return func(r *Call) {
		r.maxBytesReader = num
	}
}
