package dao

import (
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
