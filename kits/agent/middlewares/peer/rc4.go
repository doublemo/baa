package peer

import (
	"crypto/rc4"

	"github.com/doublemo/baa/kits/agent/session"
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
	first := true
	return func(next session.PeerMessageProcessor) session.PeerMessageProcessor {
		return session.PeerMessageProcessFunc(func(args session.PeerMessageProcessArgs) {
			if args.Payload.Channel == session.PeerMessageChannelWebrtc {
				next.Process(args)
				return
			}

			if encoder != nil && !first {
				encoder.XORKeyStream(args.Payload.Data, args.Payload.Data)
			}

			if first {
				first = false
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
