package dao

import (
	"strconv"

	"gorm.io/gorm"
)

const (
	defaultGroupsMaxRecord       = 10000000
	defaultGroupsMaxTable        = 50
	defaultGroupMembersMaxRecord = 10000000
	defaultGroupMembersMaxTable  = 100
)

type (
	// Groups 群组
	Groups struct {
		ID      uint64
		Name    string
		Notice  string
		Headimg string
	}

	// GroupMembers 群用户
	GroupMembers struct {
		ID          uint64
		GroupID     uint64
		UserID      uint64
		Nickname    string
		Headimg     string
		Sex         int8
		Remark      string
		Mute        int8 // 消息免打扰
		StickyOnTop int8 // 聊天置顶
		Alias       string
		Topic       uint64
	}
)

// TableName 数据库表名称
func (g Groups) TableName() string {
	return DBNamer("groups")
}

// TableName 数据库表名称
func (m GroupMembers) TableName() string {
	return DBNamer("groups", "members")
}

// UseGroupsTable 动态表名
func UseGroupsTable(id uint64) func(tx *gorm.DB) *gorm.DB {
	c32 := makeTablenameFromUint64(id, defaultGroupsMaxRecord, defaultGroupsMaxTable)
	return func(tx *gorm.DB) *gorm.DB {
		groups := &Groups{}
		tablename := DBNamer(groups.TableName(), strconv.FormatUint(uint64(c32), 10))
		_, ok := tableCacher.Get(tablename)
		if !ok {
			if !tx.Migrator().HasTable(tablename) {
				if err := tx.Table(tablename).AutoMigrate(groups); err != nil {
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

// UseGroupMembersTable 动态表名
func UseGroupMembersTable(userID uint64) func(tx *gorm.DB) *gorm.DB {
	c32 := makeTablenameFromUint64(userID, defaultGroupMembersMaxRecord, defaultGroupMembersMaxTable)
	return func(tx *gorm.DB) *gorm.DB {
		groupMembers := &GroupMembers{}
		tablename := DBNamer(groupMembers.TableName(), strconv.FormatUint(uint64(c32), 10))
		_, ok := tableCacher.Get(tablename)
		if !ok {
			if !tx.Migrator().HasTable(tablename) {
				if err := tx.Table(tablename).AutoMigrate(groupMembers); err != nil {
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
