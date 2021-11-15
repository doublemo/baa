package session

import (
	"sync"
)

var (
	defaultSessionID = "default"
)

type (
	// Session 接口
	Session interface {
		ID() string
		AddPeer(Peer)
		RemovePeer(Peer)
		Peer(string) (Peer, bool)
		Peers() []Peer
		Count() int
	}

	// SessionLocal 实现初始session
	SessionLocal struct {
		id    string
		peers map[string]Peer
		mutex sync.RWMutex
	}
)

// ID 获取session ID
func (s *SessionLocal) ID() string {
	return s.id
}

// AddPeer 添加
func (s *SessionLocal) AddPeer(peer Peer) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.peers[peer.ID()] = peer
}

// RemovePeer 删除
func (s *SessionLocal) RemovePeer(peer Peer) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.peers, peer.ID())
}

// Peer return peer in this SessionLocal
func (s *SessionLocal) Peer(id string) (peer Peer, ok bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	peer, ok = s.peers[id]
	return
}

// Peers returns peers in this SessionLocal
func (s *SessionLocal) Peers() []Peer {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	p := make([]Peer, 0, len(s.peers))
	for _, peer := range s.peers {
		p = append(p, peer)
	}
	return p
}

// Count returns peers total in this SessionLocal
func (s *SessionLocal) Count() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.peers)
}

// NewSessionLocal 创建本session
func NewSessionLocal(id string) *SessionLocal {
	return &SessionLocal{
		id:    id,
		peers: make(map[string]Peer),
	}
}

// GetSession 获取session
func GetSession(id string) (Session, bool) {
	return defaultStore.GetSession(id)
}

func RemoveSession(id string) {
	defaultStore.RemoveSessionById(id)
}

func getSession(id ...string) (Session, bool) {
	if len(id) < 1 {
		return defaultStore.GetSession(defaultSessionID)
	}

	return defaultStore.GetSession(id[0])
}

// GetPeer 获取
func GetPeer(id string, sessionId ...string) (Peer, bool) {
	session, ok := getSession(sessionId...)
	if !ok {
		return nil, false
	}

	return session.Peer(id)
}

// GetPeers 获取
func GetPeers(sessionId ...string) []Peer {
	session, ok := getSession(sessionId...)
	if !ok {
		return make([]Peer, 0)
	}

	return session.Peers()
}

// GetPeersCount 获取
func GetPeersCount(sessionId ...string) int {
	session, ok := getSession(sessionId...)
	if !ok {
		return 0
	}
	return session.Count()
}

// AddPeer 添加
func AddPeer(peer Peer, sessionId ...string) {
	session, _ := getSession(sessionId...)
	session.AddPeer(peer)
}

// RemovePeer 删除
func RemovePeer(peer Peer, sessionId ...string) {
	session, ok := getSession(sessionId...)
	if !ok {
		return
	}

	session.RemovePeer(peer)
}
