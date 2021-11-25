package interceptor

import (
	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/kits/agent/proto"
	"github.com/doublemo/baa/kits/agent/router"
	"github.com/doublemo/baa/kits/agent/session"
)

func OnStreamReceive(kitid coresproto.Command, datachannel bool) func(router.ResponseInterceptor) router.ResponseInterceptor {
	return func(next router.ResponseInterceptor) router.ResponseInterceptor {
		return router.ResponseInterceptorFunc(func(args router.ResponseInterceptorArgs) error {
			if args.Peer == nil || args.Response == nil {
				return next.Process(args)
			}

			w := proto.NewResponseBytes(kitid, args.Response)
			bytes, _ := w.Marshal()
			channel := ""
			if datachannel {
				channel = session.PeerMessageChannelWebrtc
			}

			if err := args.Peer.Send(session.PeerMessagePayload{Data: bytes, Channel: channel}); err != nil {
				return err
			}

			return next.Process(args)
		})
	}
}
