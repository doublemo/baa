package session

import "sync"

var (
	defaultStore = NewSessionStore()
)

// SessionStore session存储
type SessionStore struct {
	data  map[string]Session
	mutex sync.RWMutex
}

// AddSession 添加
func (ss *SessionStore) AddSession(session Session) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	ss.data[session.ID()] = session
}

// RemoveSession 删除
func (ss *SessionStore) RemoveSession(session Session) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	delete(ss.data, session.ID())
}

// RemoveSessionById 删除
func (ss *SessionStore) RemoveSessionById(id string) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	delete(ss.data, id)
}

// GetSession 获取
func (ss *SessionStore) GetSession(id string) (session Session, ok bool) {
	ss.mutex.RLock()
	session, ok = ss.data[id]
	ss.mutex.RUnlock()

	if !ok {
		session = NewSessionLocal(id)
	}
	return
}

// GetSessions 获取
func (ss *SessionStore) GetSessions() []Session {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	p := make([]Session, 0, len(ss.data))
	for _, session := range ss.data {
		p = append(p, session)
	}
	return p
}

// NewSessionStore new session store
func NewSessionStore() *SessionStore {
	return &SessionStore{
		data: make(map[string]Session),
	}
}
