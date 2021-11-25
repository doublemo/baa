package interceptor

import (
	"strconv"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	irouter "github.com/doublemo/baa/internal/router"
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

			if userID, ok := peer.Params("UserID"); ok {
				uid, err := strconv.ParseUint(userID.(string), 10, 64)
				if err != nil {
					return next.Process(args)
				}

				session.RemoveDict(uid, peer)
			}

			req := &corespb.Request{
				Header:  map[string]string{"PeerId": peer.ID(), "AccountID": accountID.(string)},
				Command: int32(command.AuthOffline),
			}

			req.Payload, _ = grpcproto.Marshal(&pb.Authentication_Form_Logout{
				Payload: &pb.Authentication_Form_Logout_PeerID{PeerID: peer.ID()},
			})

			_, err := mux.Handler(kit.Auth.Int32(), req)
			if err != nil {
				return err
			}

			return next.Process(args)
		})
	}
}
