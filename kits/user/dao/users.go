package dao

import (
	"strconv"

	"gorm.io/gorm"
)

const (
	defaultUsersMaxRecord = 10000000
	defaultUsersMaxTable  = 100
)

type (
	// Users 用户表
	Users struct {
		ID       uint64
		Nickname string
		Headimg  string
		Age      int8
		Sex      int8
		Idcard   string
	}
)

// TableName 数据库表名称
func (users Users) TableName() string {
	return DBNamer("users")
}

// UseUsersTable 动态表名
func UseUsersTable(userID uint64) func(tx *gorm.DB) *gorm.DB {
	c32 := makeTablenameFromUint64(userID, defaultUsersMaxRecord, defaultUsersMaxTable)
	return func(tx *gorm.DB) *gorm.DB {
		users := &Users{}
		tablename := users.TableName() + "_" + strconv.FormatUint(uint64(c32), 10)
		_, ok := tableCacher.Get(tablename)
		if !ok {
			if !tx.Migrator().HasTable(tablename) {
				if err := tx.Table(tablename).AutoMigrate(users); err != nil {
					tx.AddError(err)
				} else {
					tableCacher.Set(tablename, true, 0)
				}
			} else {
				tableCacher.Set(tablename, true, 0)
			}
		}
		return tx.Table(tablename)
	}
}
