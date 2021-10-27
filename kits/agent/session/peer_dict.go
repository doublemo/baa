package session

import (
	"sync"
	"sync/atomic"
)

var defaultDict = NewDict()

// DictLocal 用于Peer与Peer之的关系映射与查询
type DictLocal struct {
	dict       map[string]atomic.Value
	dictLen    int
	dictAllLen int
	sync.RWMutex
}

// Add 增加
func (d *DictLocal) Add(id string, peer Peer) {
	d.Lock()
	defer d.Unlock()

	m, ok := d.dict[id]
	if !ok {
		var m atomic.Value
		m.Store([]string{peer.ID()})
		d.dict[id] = m
		d.dictLen = 1
		d.dictAllLen = 1
		return
	}

	data, ok := m.Load().([]string)
	if !ok {
		var m atomic.Value
		m.Store([]string{peer.ID()})
		d.dict[id] = m
		d.dictLen = 1
		d.dictAllLen = 1
		return
	}

	dataLen := len(data)
	temp := make([]string, dataLen+1)
	tempMap := make(map[string]bool)
	for k, v := range data {
		temp[k] = v
		tempMap[v] = true
	}

	if tempMap[peer.ID()] {
		return
	}

	temp[dataLen] = peer.ID()
	m.Store(temp)
	d.dict[id] = m
	d.dictLen = len(d.dict)
	d.dictAllLen++
}

// Delete  删除
func (d *DictLocal) Delete(id string, peer Peer) {
	d.Lock()
	defer d.Unlock()

	m, ok := d.dict[id]
	if !ok {
		return
	}

	data, ok := m.Load().([]string)
	if !ok {
		return
	}

	temp := make([]string, 0)
	for _, v := range data {
		if v == peer.ID() {
			d.dictAllLen--
			continue
		}
		temp = append(temp, v)
	}

	if len(temp) < 1 {
		delete(d.dict, id)
	} else {
		m.Store(temp)
	}

	d.dict[id] = m
	d.dictLen = len(d.dict)
}

func (d *DictLocal) Get(id string) ([]string, bool) {
	d.RLock()
	m, ok := d.dict[id]
	d.RUnlock()

	if !ok {
		return nil, false
	}

	data, ok := m.Load().([]string)
	if !ok {
		return nil, false
	}

	return data, true
}

func NewDict() *DictLocal {
	return &DictLocal{
		dict: make(map[string]atomic.Value),
	}
}

func AddDict(id string, peer Peer) {
	defaultDict.Add(id, peer)
}

func RemoveDict(id string, peer Peer) {
	defaultDict.Delete(id, peer)
}

func GetDict(id string) ([]string, bool) {
	return defaultDict.Get(id)
}
