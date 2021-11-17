package peer

import (
	"crypto/rc4"

	"github.com/doublemo/baa/kits/robot/session"
)

type RC4 struct {
	encoder *rc4.Cipher
	decoder *rc4.Cipher
}

func (r *RC4) Receive() func(session.PeerMessageProcessor) session.PeerMessageProcessor {
	decoder := r.decoder
	return func(next session.PeerMessageProcessor) session.PeerMessageProcessor {
		return session.PeerMessageProcessFunc(func(args session.PeerMessageProcessArgs) {
			if args.Payload.Channel == session.PeerMessageChannelWebrtc {
				next.Process(args)
				return
			}

			if decoder != nil {
				decoder.XORKeyStream(args.Payload.Data, args.Payload.Data)
			}
			next.Process(args)
		})
	}
}

func (r *RC4) Write() func(session.PeerMessageProcessor) session.PeerMessageProcessor {
	encoder := r.encoder
	return func(next session.PeerMessageProcessor) session.PeerMessageProcessor {
		return session.PeerMessageProcessFunc(func(args session.PeerMessageProcessArgs) {
			if args.Payload.Channel == session.PeerMessageChannelWebrtc {
				next.Process(args)
				return
			}

			if encoder != nil {
				encoder.XORKeyStream(args.Payload.Data, args.Payload.Data)
			}

			next.Process(args)
		})
	}
}

func NewRC4(encoder, decoder *rc4.Cipher) *RC4 {
	return &RC4{
		encoder: encoder,
		decoder: decoder,
	}
}
