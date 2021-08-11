package rpc

import (
	"context"
	"io"
	"sync"

	logger "github.com/doublemo/baa/cores/log"
	log "github.com/doublemo/baa/cores/log/level"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type (
	// BidirectionalStreamingClient 创建流客户端
	BidirectionalStreamingClient struct {
		OnReceive func(*corespb.Response)
		OnClose   func()
		conn      *grpc.ClientConn
		stream    corespb.Service_BidirectionalStreamingClient
		quitCh    chan struct{}
		once      sync.Once
		logger    logger.Logger
		mutex     sync.Mutex
	}
)

func (bsc *BidirectionalStreamingClient) Connect(md metadata.MD) error {
	client := corespb.NewServiceClient(bsc.conn)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	stream, err := client.BidirectionalStreaming(ctx)
	if err != nil {

		return err
	}

	bsc.stream = stream

	readyedChan := make(chan struct{})
	dataChan := make(chan *corespb.Response, 1)
	go bsc.recv(dataChan, readyedChan)
	go bsc.recvForStream(dataChan, readyedChan)
	<-readyedChan
	<-readyedChan
	close(readyedChan)
	return nil
}

func (bsc *BidirectionalStreamingClient) recvForStream(dataChan chan *corespb.Response, readyedChan chan struct{}) {
	defer close(dataChan)

	readyedChan <- struct{}{}
	for {
		frame, err := bsc.stream.Recv()
		if err == io.EOF {
			return
		}

		if err != nil {
			code := status.Code(err)
			if code != codes.Canceled {
				log.Error(bsc.logger).Log("stream.Recv() returned expected error", err)
			}
			return
		}

		dataChan <- frame
	}
}

func (bsc *BidirectionalStreamingClient) recv(dataChan chan *corespb.Response, readyedChan chan struct{}) {
	defer func() {
		bsc.mutex.Lock()
		bsc.stream.CloseSend()
		bsc.mutex.Unlock()

		if bsc.OnClose != nil {
			bsc.OnClose()
		}
	}()

	readyedChan <- struct{}{}
	for {
		select {
		case frame, ok := <-dataChan:
			if !ok {
				return
			}

			if bsc.OnReceive != nil {
				bsc.OnReceive(frame)
			}

		case <-bsc.quitCh:
			return
		}
	}

}

func (bsc *BidirectionalStreamingClient) Send(r *corespb.Request) error {
	bsc.mutex.Lock()
	defer bsc.mutex.Unlock()
	return bsc.stream.Send(r)
}

func (bsc *BidirectionalStreamingClient) Close() {
	bsc.once.Do(func() {
		close(bsc.quitCh)
	})
}

// NewBidirectionalStreamingClient 创建流客户端
func NewBidirectionalStreamingClient(conn *grpc.ClientConn, l logger.Logger) *BidirectionalStreamingClient {
	return &BidirectionalStreamingClient{
		conn:   conn,
		quitCh: make(chan struct{}),
		logger: l,
	}
}
