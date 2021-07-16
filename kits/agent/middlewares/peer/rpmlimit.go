package peer

import (
	"fmt"
	"time"

	"github.com/doublemo/baa/cores/log"
	kitlog "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/kits/agent/session"
)

type RPMLimiter struct {
	rpmlimt int
	logger  log.Logger
}

func (rpm *RPMLimiter) Receive() func(session.PeerMessageProcessor) session.PeerMessageProcessor {
	counter := 0
	last := time.Now()
	rpmLimit := rpm.rpmlimt
	logger := rpm.logger
	return func(next session.PeerMessageProcessor) session.PeerMessageProcessor {
		return session.PeerMessageProcessFunc(func(args session.PeerMessageProcessArgs) {
			counter++
			if time.Now().Sub(last).Seconds() > 60 {
				last = time.Now()
				counter = 0
			}

			if counter > rpmLimit {
				kitlog.Warn(logger).Log("Warn", fmt.Sprintf("The Peer is rpmlimt greater than the maximum threshold"))
				args.Peer.Close()
				return
			}

			next.Process(args)
		})
	}
}

func (rpm *RPMLimiter) Write() func(session.PeerMessageProcessor) session.PeerMessageProcessor {
	return nil
}

// NewRPMLimiter new rpm limiter
func NewRPMLimiter(limt int, logger log.Logger) *RPMLimiter {
	return &RPMLimiter{
		rpmlimt: limt,
		logger:  logger,
	}
}
