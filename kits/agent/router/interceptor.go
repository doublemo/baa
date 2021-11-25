package router

import (
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/kits/agent/session"
)

type (
	// RequestInterceptorArgs 拦截器参数
	RequestInterceptorArgs struct {
		Peer          session.Peer
		Request       *corespb.Request
		ClientRequest coresproto.Request
	}

	// ResponseInterceptorArgs 拦截器参数
	ResponseInterceptorArgs struct {
		Peer     session.Peer
		Response *corespb.Response
	}

	// RequestInterceptor 输入拦截器
	RequestInterceptor interface {
		Process(args RequestInterceptorArgs) error
	}

	RequestInterceptorFunc func(args RequestInterceptorArgs) error

	// ResponseInterceptor 输出拦截器
	ResponseInterceptor interface {
		Process(args ResponseInterceptorArgs) error
	}

	ResponseInterceptorFunc func(args ResponseInterceptorArgs) error

	RequestInterceptors []func(RequestInterceptor) RequestInterceptor

	ResponseInterceptors []func(ResponseInterceptor) ResponseInterceptor

	chainRequestInterceptor struct {
		interceptors RequestInterceptors
		last         RequestInterceptor
		current      RequestInterceptor
	}

	chainResponseInterceptor struct {
		interceptors ResponseInterceptors
		last         ResponseInterceptor
		current      ResponseInterceptor
	}
)

func (c chainRequestInterceptor) Process(args RequestInterceptorArgs) error {
	return c.current.Process(args)
}

func (c chainResponseInterceptor) Process(args ResponseInterceptorArgs) error {
	return c.current.Process(args)
}

func (interceptors RequestInterceptors) Process(interceptor RequestInterceptor) RequestInterceptor {
	return &chainRequestInterceptor{interceptors, interceptor, chainRequest(interceptors, interceptor)}
}

func (interceptors RequestInterceptors) ProcessFunc(interceptor RequestInterceptor) RequestInterceptor {
	return &chainRequestInterceptor{interceptors, interceptor, chainRequest(interceptors, interceptor)}
}

func (interceptor RequestInterceptorFunc) Process(args RequestInterceptorArgs) error {
	return interceptor(args)
}

func (interceptors ResponseInterceptors) Process(interceptor ResponseInterceptor) ResponseInterceptor {
	return &chainResponseInterceptor{interceptors, interceptor, chainResponse(interceptors, interceptor)}
}

func (interceptors ResponseInterceptors) ProcessFunc(interceptor ResponseInterceptor) ResponseInterceptor {
	return &chainResponseInterceptor{interceptors, interceptor, chainResponse(interceptors, interceptor)}
}

func (interceptor ResponseInterceptorFunc) Process(args ResponseInterceptorArgs) error {
	return interceptor(args)
}

func chainRequest(interceptors []func(processor RequestInterceptor) RequestInterceptor, last RequestInterceptor) RequestInterceptor {
	if len(interceptors) == 0 {
		return last
	}
	h := interceptors[len(interceptors)-1](last)
	for i := len(interceptors) - 2; i >= 0; i-- {
		h = interceptors[i](h)
	}
	return h
}

func chainResponse(interceptors []func(processor ResponseInterceptor) ResponseInterceptor, last ResponseInterceptor) ResponseInterceptor {
	if len(interceptors) == 0 {
		return last
	}
	h := interceptors[len(interceptors)-1](last)
	for i := len(interceptors) - 2; i >= 0; i-- {
		h = interceptors[i](h)
	}
	return h
}

func WithRequestInterceptor(interceptors ...func(RequestInterceptor) RequestInterceptor) RequestInterceptors {
	return RequestInterceptors(interceptors)
}

func WithResponseInterceptor(interceptors ...func(ResponseInterceptor) ResponseInterceptor) ResponseInterceptors {
	return ResponseInterceptors(interceptors)
}
