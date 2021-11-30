package interceptor

import (
	"fmt"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	irouter "github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/agent/router"
	"github.com/doublemo/baa/kits/agent/session"
	grpcproto "github.com/golang/protobuf/proto"
)

func OnOfflineRouterDestroy(mux *irouter.Mux) func(router.ResponseInterceptor) router.ResponseInterceptor {
	return func(next router.ResponseInterceptor) router.ResponseInterceptor {
		return router.ResponseInterceptorFunc(func(args router.ResponseInterceptorArgs) error {
			if args.Peer == nil {
				return next.Process(args)
			}

			peer := args.Peer
			accountID, ok := peer.Params("AccountID")
			if !ok {
				return next.Process(args)
			}

			userID, ok := peer.Params("UserID")
			if !ok {
				return next.Process(args)
			}

			session.RemoveDict(userID.(string), peer)
			req := &corespb.Request{
				Header:  map[string]string{"PeerId": peer.ID()},
				Command: int32(command.AuthOffline),
			}

			req.Payload, _ = grpcproto.Marshal(&pb.Authentication_Form_Logout{
				AccountID: accountID.(string),
				UserID:    userID.(string),
			})

			_, err := mux.Handler(kit.Auth.Int32(), req)
			if err != nil {
				return err
			}

			return next.Process(args)
		})
	}
}

// OnSelectIMServer 选择聊天服务器
func OnSelectIMServer(next router.RequestInterceptor) router.RequestInterceptor {
	return router.RequestInterceptorFunc(func(args router.RequestInterceptorArgs) error {
		if args.Peer == nil || args.Request == nil {
			return next.Process(args)
		}

		peer := args.Peer
		imhost, ok := peer.Params("IMServer-Host-Addr")
		if ok {
			args.Request.Header["Host"] = imhost.(string)
			return next.Process(args)
		}

		im, ok := peer.Params("IMServer")
		if !ok {
			return next.Process(args)
		}

		endponts, err := sd.GetEndpointsByID(im.(string))
		if err != nil {
			return fmt.Errorf("IM server is undefined, %s", err.Error())
		}

		imserverEndpont, ok := endponts[im.(string)]
		if err != nil {
			return fmt.Errorf("IM server is undefined, %s", im)
		}

		peer.SetParams("IMServer-Host-Addr", imserverEndpont.Addr())
		args.Request.Header["Host"] = imserverEndpont.Addr()
		return next.Process(args)
	})
}

func AddIMServerToHeader(next router.RequestInterceptor) router.RequestInterceptor {
	return router.RequestInterceptorFunc(func(args router.RequestInterceptorArgs) error {
		if args.Peer == nil || args.Request == nil {
			return next.Process(args)
		}

		peer := args.Peer
		imhost, ok := peer.Params("IMServer-Host-Addr")
		if ok {
			args.Request.Header["IMServer"] = imhost.(string)
			return next.Process(args)
		}

		im, ok := peer.Params("IMServer")
		if !ok {
			return next.Process(args)
		}

		endponts, err := sd.GetEndpointsByID(im.(string))
		if err != nil {
			return fmt.Errorf("IM server is undefined, %s", err.Error())
		}

		imserverEndpont, ok := endponts[im.(string)]
		if err != nil {
			return fmt.Errorf("IM server is undefined, %s", im)
		}

		peer.SetParams("IMServer-Host-Addr", imserverEndpont.Addr())
		args.Request.Header["IMServer"] = imserverEndpont.Addr()
		return next.Process(args)
	})
}
