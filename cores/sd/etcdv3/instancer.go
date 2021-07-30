package etcdv3

import (
	"fmt"

	"github.com/doublemo/baa/cores/sd"
	"github.com/doublemo/baa/cores/sd/internal/instance"
)

// Instancer yields instances stored in a certain etcd keyspace. Any kind of
// change in that keyspace is watched and will update the Instancer's Instancers.
type Instancer struct {
	cache  *instance.Cache
	client Client
	prefix string
	quitc  chan struct{}
}

// NewInstancer returns an etcd instancer. It will start watching the given
// prefix for changes, and update the subscribers.
func NewInstancer(c Client, prefix string) (*Instancer, error) {
	s := &Instancer{
		cache:  instance.NewCache(),
		client: c,
		prefix: prefix,
		quitc:  make(chan struct{}),
	}

	instances, err := s.client.GetEntries(s.prefix)
	if err != nil {
		fmt.Println("err:", err)
	}

	s.cache.Update(sd.Event{Instances: instances, Err: err})
	go s.loop()
	return s, nil
}

func (s *Instancer) loop() {
	defer func() {
		fmt.Println("loop return")
	}()
	ch := make(chan struct{})
	go s.client.WatchPrefix(s.prefix, ch)

	for {
		select {
		case <-ch:
			instances, err := s.client.GetEntries(s.prefix)
			if err != nil {
				fmt.Println("msg", "failed to retrieve entries", "err", err)
				s.cache.Update(sd.Event{Err: err})
				continue
			}

			s.cache.Update(sd.Event{Instances: instances})

		case <-s.quitc:
			return
		}
	}
}

// Stop terminates the Instancer.
func (s *Instancer) Stop() {
	close(s.quitc)
}

// Register implements Instancer.
func (s *Instancer) Register(ch chan<- sd.Event) {
	s.cache.Register(ch)
}

// Deregister implements Instancer.
func (s *Instancer) Deregister(ch chan<- sd.Event) {
	s.cache.Deregister(ch)
}
