package dao

import (
	"strconv"

	"gorm.io/gorm"
)

const (
	defaultContactsMaxRecord = 10000000
	defaultContactsMaxTable  = 100
)

// Contacts 通讯录
type Contacts struct {
	ID          uint64
	UserID      uint64
	FriendID    uint64
	FNickname   string
	FHeadimg    string
	FSex        int8
	Remark      string // 备注
	Mute        int8   // 消息免打扰
	StickyOnTop int8   // 聊天置顶
	Type        int8   // 好友类型
	Topic       uint64 // crc64
}

// TableName 数据库表名称
func (contacts Contacts) TableName() string {
	return DBNamer("users", "contacts")
}

// UseContactsTable 动态联系人表名
func UseContactsTable(userID uint64) func(tx *gorm.DB) *gorm.DB {
	c32 := makeTablenameFromUint64(userID, defaultContactsMaxRecord, defaultContactsMaxTable)
	return func(tx *gorm.DB) *gorm.DB {
		contacts := &Contacts{}
		tablename := DBNamer(contacts.TableName(), strconv.FormatUint(uint64(c32), 10))
		_, ok := tableCacher.Get(tablename)
		if !ok {
			if !tx.Migrator().HasTable(tablename) {
				if err := tx.Table(tablename).AutoMigrate(contacts); err != nil {
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
