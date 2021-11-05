package sm

import (
	"fmt"
	"math/rand"
	"sync"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/internal/sd"
)

// Endpoints 节点管理
type Endpoints struct {
	records    map[string][]string
	wch        chan struct{}
	closeCh    chan struct{}
	r          *rand.Rand
	roundRobin map[string]int
	mutex      sync.RWMutex
	rrmutex    sync.Mutex
	once       sync.Once
}

// Random 随机获取节点
func (endpoints *Endpoints) Random(name string) (string, bool) {
	var servers []string
	endpoints.mutex.RLock()
	if m, ok := endpoints.records[name]; ok && len(m) > 0 {
		servers = make([]string, len(m))
		copy(servers[0:], m[0:])
	}
	endpoints.mutex.RUnlock()
	if len(servers) < 1 {
		return "", false
	}
	return servers[endpoints.r.Intn(len(servers))], true
}

// RoundRobin 循环
func (endpoints *Endpoints) RoundRobin(name string) (string, bool) {
	var servers []string
	endpoints.mutex.RLock()
	if m, ok := endpoints.records[name]; ok && len(m) > 0 {
		servers = make([]string, len(m))
		copy(servers[0:], m[0:])
	}
	endpoints.mutex.RUnlock()
	if len(servers) < 1 {
		return "", false
	}

	endpoints.rrmutex.Lock()
	rbindx := endpoints.roundRobin[name]
	rbindx++
	if rbindx >= len(servers) {
		rbindx = 0
	}
	endpoints.roundRobin[name] = rbindx
	endpoints.rrmutex.Unlock()
	return servers[rbindx], true
}

func (endpoints *Endpoints) Endpoints(name string) []string {
	var servers []string
	endpoints.mutex.RLock()
	if m, ok := endpoints.records[name]; ok && len(m) > 0 {
		servers = make([]string, len(m))
		copy(servers[0:], m[0:])
	}
	endpoints.mutex.RUnlock()
	return servers
}

// Watch 节点信息监控
func (endpoints *Endpoints) Watch() {
	for {
		select {
		case <-endpoints.wch:
			endpoints.reload()

		case <-endpoints.closeCh:
			return
		}
	}
}

// Close 关闭
func (endpoints *Endpoints) Close() {
	endpoints.once.Do(func() {
		close(endpoints.closeCh)
	})
}

func (endpoints *Endpoints) reload() {
	eds, err := sd.Endpoints()
	if err != nil {
		log.Error(Logger()).Log("action", "reload", "error", err)
		return
	}

	data := make(map[string][]string)
	for _, e := range eds {
		if _, ok := data[e.Name()]; !ok {
			data[e.Name()] = make([]string, 0)
		}
		data[e.Name()] = append(data[e.Name()], e.ID())
	}

	endpoints.mutex.Lock()
	endpoints.records = data
	endpoints.mutex.Unlock()
	fmt.Println(endpoints.Random("auth"))
	fmt.Println(endpoints.RoundRobin("im"))
	fmt.Println(endpoints.RoundRobin("user"))
}

func NewEndpoints(wch chan struct{}, seed int64) *Endpoints {
	return &Endpoints{
		records:    make(map[string][]string),
		roundRobin: make(map[string]int),
		wch:        wch,
		r:          rand.New(rand.NewSource(seed)),
		closeCh:    make(chan struct{}),
	}
}
