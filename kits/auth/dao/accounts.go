package dao

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Accounts 账户
type Accounts struct {
	ID        uint64 `gorm:"<-:create;primaryKey"`
	UnionID   uint64 `gorm:"<-:create;index"`
	UserID    uint64 `gorm:"<-:create;index"`
	Scheme    string `gorm:"<-:create;size:50;index:scheme_name"`
	Name      string `gorm:"<-:create;index:scheme_name"`
	Secret    string
	Status    int
	ExpiresAt int64
	PeerID    string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// GetAccoutsBySchemeAName 根据条件 scheme, name 获取信息
func GetAccoutsBySchemeAName(scheme, name string) (*Accounts, error) {
	if db == nil {
		return nil, gorm.ErrInvalidDB
	}

	accounts := &Accounts{}
	tx := db.Where("scheme = ? AND name = ?", scheme, name).First(accounts)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return accounts, nil
}

func GetAccoutsByPeerID(id string) (*Accounts, error) {
	if db == nil {
		return nil, gorm.ErrInvalidDB
	}

	accounts := &Accounts{}
	tx := db.Where("peer_id = ?", id).First(accounts)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return accounts, nil
}

func CreateAccount(accounts *Accounts) error {
	if db == nil {
		return gorm.ErrInvalidDB
	}

	r := db.Create(accounts)
	if r.Error != nil {
		return r.Error
	}

	if r.RowsAffected != 1 {
		return errors.New("CreateFailed")
	}

	return nil
}

func UpdatesAccountByID(id uint64, col string, value interface{}) (int64, error) {
	if db == nil {
		return 0, gorm.ErrInvalidDB
	}

	r := db.Model(&Accounts{}).Where("id = ?", id).Update(col, value)
	return r.RowsAffected, r.Error
}
