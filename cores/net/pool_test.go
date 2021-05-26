// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/baa>

package net

import (
	"math"
	"testing"
	"time"

	"sync/atomic"
)

type TestPoolAdapter struct {
	id int
}

func (tpa *TestPoolAdapter) Close() {}

func (tpa *TestPoolAdapter) Ok() bool {
	return true
}

func TestConnPool(t *testing.T) {
	c := NewConnPool(1, 500, time.Second*60)
	var (
		m              int32
		counterSuccess int32
		counterGetFail int32
		threadNum      = 1000
	)

	c.New(func() (PoolAdapter, error) {
		atomic.AddInt32(&m, 1)
		return &TestPoolAdapter{id: int(atomic.LoadInt32(&m))}, nil
	})

	for n := 0; n < threadNum; n++ {
		//time.Sleep(time.Millisecond * 10)
		go func(o int) {
			for i := 0; i < 1; i++ {
				//begin := time.Now()
				x, err := c.Get()
				//end := time.Now().Sub(begin)

				if err != nil {
					atomic.AddInt32(&counterGetFail, 1)
					//t.Fatal(err, end.Seconds(), "s")
					return
				}

				time.Sleep(time.Millisecond * 100)
				c.Put(x)
				atomic.AddInt32(&counterSuccess, 1)
			}
		}(n)
	}

	time.Sleep(time.Second * 10)
	n := math.Trunc((float64(atomic.LoadInt32(&counterSuccess))/float64(threadNum))*1e2+0.5) * 1e-2
	t.Log("Len", c.len(), "MaxUD", m, "Success", atomic.LoadInt32(&counterSuccess), "Failed", atomic.LoadInt32(&counterGetFail), "Sum", n*100, "%", c.counter)
}
