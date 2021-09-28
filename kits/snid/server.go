package snid

import (
	"context"
	"fmt"

	corespb "github.com/doublemo/baa/cores/proto/pb"
)

type baseServer struct {
	corespb.UnimplementedServiceServer
}

func (s *baseServer) Call(ctx context.Context, req *corespb.Request) (*corespb.Response, error) {
	fmt.Println("call:", req)
	return r.Handler(req)
}

func (s *baseServer) BidirectionalStreaming(stream corespb.Service_BidirectionalStreamingServer) error {
	return nil
}
