package helper

import (
	"math/rand"
	"time"

	"github.com/doublemo/baa/cores/crypto/token"
)

// Token token
type Token struct {
	// ID 账户ID
	ID uint64

	// UnionID 联合ID
	UnionID uint64

	// PeerID 网关PeerID
	PeerID string

	// Expires session过期时间
	Expires int64

	T int32
}

// GenerateToken 创建token string
func GenerateToken(id, unionID uint64, peerID string, expireat time.Duration, secret []byte) (string, error) {
	tk := token.NewTK(secret)
	expires := time.Now().Add(expireat)
	return tk.Encrypt(&Token{ID: id, UnionID: unionID, PeerID: peerID, Expires: expires.Unix(), T: rand.Int31()})
}

// ParseToken 解析token
func ParseToken(s string, secret []byte) (*Token, error) {
	data := &Token{}
	tk := token.NewTK(secret)
	if err := tk.Decrypt(s, data); err != nil {
		return nil, err
	}

	return data, nil
}

// IsValidToken 检查token是否有效
func IsValidToken(t *Token) bool {
	if t.Expires < time.Now().Unix() {
		return false
	}

	return true
}
