package session

import (
	"sync"
	"testing"
	"time"
)

func TestDict(t *testing.T) {
	d := NewDict()
	peers := make([]Peer, 0)
	for i := 0; i < 1000; i++ {
		peers = append(peers, NewPeerSocket(nil, 0, 0, make(chan struct{})))
	}

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		for _, p := range peers {
			d.Add("test", p)
		}
	}()

	go func() {
		defer wg.Done()
		for _, p := range peers {
			d.Delete("test", p)
			time.Sleep(time.Millisecond)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			d.Get("test")
		}
	}()
	wg.Wait()
	t.Log(d.Get("test"))
	t.Log("ok", d.dictAllLen, d.dictLen)
}
