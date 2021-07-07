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

// const (
// 	RC4EncoderKey  = "rc4encoder"
// 	RC4DecoderKey  = "rc4decoder"
// 	RC4NoEncodeKey = "rc4noencode"
// 	RC4NoDecodeKey = "rc4nodecode"
// )
// func (rpm *RC4) Receive() func(session.PeerMessageProcessor) session.PeerMessageProcessor {
// 	return func(next session.PeerMessageProcessor) session.PeerMessageProcessor {
// 		return session.PeerMessageProcessFunc(func(args session.PeerMessageProcessArgs) {
// 			peer := args.Peer
// 			if rc4decoder, ok := peer.Params(RC4DecoderKey); ok && rc4decoder != nil {
// 				nodecode, noexit := peer.Params(RC4NoDecodeKey)
// 				if noexit {
// 					if nodecode != nil {
// 						peer.SetParams(RC4NoDecodeKey, nil)
// 						next.Process(args)
// 						return
// 					}
// 				}

// 				decoder, ok := rc4decoder.(*rc4.Cipher)
// 				if ok {
// 					frame := args.Payload.Data
// 					decoder.XORKeyStream(frame, frame)
// 					args.Payload.Data = frame
// 				}
// 			}
// 			next.Process(args)
// 		})
// 	}
// }

// func (rpm *RC4) Write() func(session.PeerMessageProcessor) session.PeerMessageProcessor {
// 	return func(next session.PeerMessageProcessor) session.PeerMessageProcessor {
// 		return session.PeerMessageProcessFunc(func(args session.PeerMessageProcessArgs) {
// 			peer := args.Peer
// 			if rc4encoder, ok := peer.Params(RC4EncoderKey); ok && rc4encoder != nil {
// 				noencode, noexit := peer.Params(RC4NoEncodeKey)
// 				if noexit {
// 					if noencode != nil {
// 						peer.SetParams(RC4NoEncodeKey, nil)
// 						next.Process(args)
// 						return
// 					}
// 				}

// 				encoder, ok := rc4encoder.(*rc4.Cipher)
// 				if ok {
// 					frame := args.Payload.Data
// 					encoder.XORKeyStream(frame, frame)
// 					args.Payload.Data = frame
// 				}
// 			}
// 			next.Process(args)
// 		})
// 	}
// }
