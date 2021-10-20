package queue

import (
	"container/heap"
	"sync"
)

type (
	// orderedUint64Item 有序uint64队列
	orderedUint64Item struct {
		value uint64
		index int
	}

	orderedUint64Queue []*orderedUint64Item

	// OrderedUint64 有序uint64队列
	OrderedUint64 struct {
		queue orderedUint64Queue
		sync.RWMutex
	}
)

// Len 长度
func (pq orderedUint64Queue) Len() int { return len(pq) }

// Less 策略排序比较
func (pq orderedUint64Queue) Less(i, j int) bool {
	return pq[i].value < pq[j].value
}

// Swap 交换
func (pq orderedUint64Queue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push 写入
func (pq *orderedUint64Queue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*orderedUint64Item)
	item.index = n
	*pq = append(*pq, item)
}

// Pop 弹出
func (pq *orderedUint64Queue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func (pq *orderedUint64Queue) update(item *priorityItem, value interface{}, priority int) {
	item.value = value
	item.priority = priority
	heap.Fix(pq, item.index)
}

// Pop 弹出
func (p *OrderedUint64) Pop() uint64 {
	p.RLock()
	defer p.RUnlock()
	if p.queue.Len() < 1 {
		return 0
	}

	item := heap.Pop(&p.queue).(*orderedUint64Item)
	return item.value
}

// Push 写入
func (p *OrderedUint64) Push(values ...uint64) {
	p.Lock()
	defer p.Unlock()
	for _, value := range values {
		heap.Push(&p.queue, &orderedUint64Item{value: value})
	}
}

// Len 长度
func (p *OrderedUint64) Len() int {
	p.RLock()
	defer p.RUnlock()
	return p.queue.Len()
}

// NewOrderedUint64 new
func NewOrderedUint64(values ...uint64) *OrderedUint64 {
	pr := &OrderedUint64{queue: make(orderedUint64Queue, len(values))}
	for idx, value := range values {
		pr.queue[idx] = &orderedUint64Item{value: value, index: idx}
	}

	heap.Init(&pr.queue)
	return pr
}
