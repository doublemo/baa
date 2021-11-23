// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

package os

import (
	"sort"
	"sync"
	"sync/atomic"
)

type ProcessActor struct {
	id int32

	stoped int32

	Exec func() error

	Interrupt func(error)

	Close func()
}

type processError struct {
	id  int32
	err error
}

// Process 协程管理盒子
type Process struct {
	actors  sync.Map
	counter int32
}

// Add 增加协程服务到盒子
func (rc *Process) Add(actor *ProcessActor, status bool) int32 {
	if actor == nil || actor.Exec == nil {
		return 0
	}

	actor.id = atomic.AddInt32(&rc.counter, 1)
	if status {
		actor.stoped = 1
	} else {
		actor.stoped = 0
	}

	rc.actors.Store(actor.id, actor)
	return actor.id
}

// Run 运行盒子内
func (rc *Process) Run() error {
	errors := make(chan processError)
	defer close(errors)

	actors := rc.sortProcess(1)
	waitCounter := 0
	for _, actor := range actors {
		if atomic.LoadInt32(&actor.stoped) == 0 || actor == nil || actor.Exec == nil {
			continue
		}

		waitCounter++
		atomic.StoreInt32(&actor.stoped, 0)
		go func(a *ProcessActor) {
			errors <- processError{id: a.id, err: a.Exec()}
		}(actor)
	}

	if waitCounter < 1 {
		return nil
	}

	var (
		doneCounter int
		err         error
	)

	for e := range errors {
		doneCounter++
		if m, ok := rc.actors.Load(e.id); ok {
			actor := m.(*ProcessActor)
			if atomic.LoadInt32(&actor.stoped) == 0 {
				if actor.Interrupt != nil {
					actor.Interrupt(e.err)
				}
				atomic.StoreInt32(&actor.stoped, 1)
			}
		}

		if err == nil && e.err != nil {
			err = e.err
		}

		if doneCounter >= waitCounter {
			break
		}
	}
	return err
}

// Stop 关闭盒子内所有服务
func (rc *Process) Stop() {
	actors := rc.sortProcess(0)
	for i := len(actors) - 1; i >= 0; i-- {
		actor := (*ProcessActor)(actors[i])
		if atomic.LoadInt32(&actor.stoped) == 1 {
			continue
		}

		if actor != nil {
			actor.Close()
		}
	}
}

func (rc *Process) sortProcess(status int32) []*ProcessActor {
	data := make([]*ProcessActor, 0)
	rc.actors.Range(func(k, v interface{}) bool {
		actor := v.(*ProcessActor)
		if atomic.LoadInt32(&actor.stoped) == status {
			data = append(data, actor)
		}

		return true
	})

	sort.Slice(data, func(a, b int) bool {
		actorA := (*ProcessActor)(data[a])
		actorB := (*ProcessActor)(data[b])
		if actorA.id > actorB.id {
			return false
		}
		return true
	})

	return data
}

// NewProcess 创建盒子
func NewProcess() *Process {
	c := &Process{}
	return c
}
