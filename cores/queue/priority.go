package queue

import (
	"container/heap"
	"sync"
)

type (
	// priorityItem 优先级队列类型
	priorityItem struct {
		value    interface{}
		priority int
		index    int
	}

	// priorityQueue 优先级队列
	priorityQueue []*priorityItem

	// Priority heap优先级队列
	Priority struct {
		queue priorityQueue
		sync.RWMutex
	}
)

// Len 长度
func (pq priorityQueue) Len() int { return len(pq) }

// Less 策略排序比较
func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].priority > pq[j].priority
}

// Swap 交换
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push 写入
func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*priorityItem)
	item.index = n
	*pq = append(*pq, item)
}

// Pop 弹出
func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func (pq *priorityQueue) update(item *priorityItem, value interface{}, priority int) {
	item.value = value
	item.priority = priority
	heap.Fix(pq, item.index)
}

// Pop 弹出
func (p *Priority) Pop() interface{} {
	p.RLock()
	defer p.RUnlock()
	if p.queue.Len() < 1 {
		return nil
	}

	item := heap.Pop(&p.queue).(*priorityItem)
	return item.value
}

// Push 写入
func (p *Priority) Push(priority int, value interface{}) {
	p.Lock()
	defer p.Unlock()

	heap.Push(&p.queue, &priorityItem{value: value, priority: priority})
}

func (p *Priority) Len() int {
	p.RLock()
	defer p.RUnlock()
	return p.queue.Len()
}

// NewPriority new
func NewPriority(values map[int]interface{}) *Priority {
	pr := &Priority{queue: make(priorityQueue, len(values))}
	idx := 0
	for priority, value := range values {
		pr.queue[idx] = &priorityItem{value: value, priority: priority, index: idx}
		idx++
	}

	heap.Init(&pr.queue)
	return pr
}
