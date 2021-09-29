package dao

import (
	"gorm.io/gorm"
)

// Accounts 账户
type Accounts struct {
	gorm.Model
	UnionID   uint64 `gorm:"<-:create;uniqueIndex"`
	Scheme    string `gorm:"<-:create;size:50;index:scheme_name"`
	Name      string `gorm:"<-:create;index:scheme_name"`
	Secret    string
	Status    int
	ExpiresAt int64
}
