package dao

import (
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
	return useTable(id, &Groups{}, defaultGroupsMaxRecord, defaultGroupsMaxTable)
}

// UseGroupMembersTable 动态表名
func UseGroupMembersTable(userID uint64) func(tx *gorm.DB) *gorm.DB {
	return useTable(userID, &GroupMembers{}, defaultGroupMembersMaxRecord, defaultGroupMembersMaxTable)
}
