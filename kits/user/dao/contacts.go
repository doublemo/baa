package dao

import (
	"errors"
	"time"

	"github.com/doublemo/baa/internal/helper"
	"gorm.io/gorm"
)

const (
	defaultContactsMaxRecord        = 10000000
	defaultContactsMaxTable         = 100
	defaultContactsRequestMaxRecord = 10000000
	defaultContactsRequestMaxTable  = 10
)

const (
	// ContactsTypePerson 个人
	ContactsTypePerson = 1

	// ContactsTypeGroup 群
	ContactsTypeGroup = 2
)

type (
	// Contacts 通讯录
	Contacts struct {
		ID          uint64 `gorm:"<-:create;primaryKey;autoIncrement"`
		UserID      uint64 `gorm:"<-:create;index;index:userid_friendid;priority:1"`
		FriendID    uint64 `gorm:"<-:create;index:userid_friendid;priority:2"`
		FNickname   string `gorm:"size:50"`
		FHeadimg    string `gorm:"size:256"`
		FSex        int8
		Remark      string `gorm:"size:50"` // 备注
		Mute        int8   // 消息免打扰
		StickyOnTop int8   // 聊天置顶
		Type        int8   // 好友类型
		Topic       uint64 `gorm:"index"` // crc32
		Status      int8   // 好友状态 0 正常 1 拉黑
		Version     int64  // 好友信息更新版本号
		CreateAt    int64  `gorm:"autoCreateTime"`
	}

	// ContactsRequest 增加联系人请
	ContactsRequest struct {
		ID        uint64 `gorm:"<-:create;primaryKey;autoIncrement"`
		UserID    uint64 `gorm:"<-:create;index;index:userid_friendid;priority:1"`
		FriendID  uint64 `gorm:"<-:create;index;index:userid_friendid;priority:2"`
		FromID    uint64
		FNickname string `gorm:"size:50"`
		FHeadimg  string `gorm:"size:256"`
		FSex      int8
		Remark    string `gorm:"size:50"`    // 备注
		Messages  string `gorm:"size:10000"` // size消息
		Status    int8   // 状态  -1 已经过期 0 正常 1 拒绝
		Version   int64  // 好友信息更新版本号
		From      string // 来源
		CreatedAt int64  `gorm:"autoCreateTime"`
	}
)

// TableName 数据库表名称
func (contacts Contacts) TableName() string {
	return DBNamer("users", "contacts")
}

// TableName 数据库表名称
func (r ContactsRequest) TableName() string {
	return DBNamer("users", "contacts", "request")
}

// UseContactsTable 动态表名
func UseContactsTable(userID uint64) func(tx *gorm.DB) *gorm.DB {
	return useTable(userID, &Contacts{}, defaultContactsMaxRecord, defaultContactsMaxTable)
}

// UseContactsRequestTable 动态表名
func UseContactsRequestTable(userID uint64) func(tx *gorm.DB) *gorm.DB {
	return useTable(userID, &ContactsRequest{}, defaultContactsRequestMaxRecord, defaultContactsRequestMaxTable)
}

// FindContactsByUserIDAndFriendID 查询是否已经为好友
func FindContactsByUserIDAndFriendID(userid, friendid uint64, cols ...string) (*Contacts, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	contacts := &Contacts{}
	tx := database.Scopes(UseContactsTable(userid))
	if len(cols) > 0 {
		tx.Select(cols)
	}

	tx.Where("user_id = ? And friend_id = ?", userid, friendid).Take(contacts)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}

		return nil, tx.Error
	}

	return contacts, nil
}

// FindContactsByUserIDAndTopic 查询是否已经为好友
func FindContactsByUserIDAndTopic(userid, topic uint64, cols ...string) (*Contacts, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	contacts := &Contacts{}
	tx := database.Scopes(UseContactsTable(userid))
	if len(cols) > 0 {
		tx.Select(cols)
	}

	tx.Where("topic = ?", topic).Take(contacts)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}

		return nil, tx.Error
	}

	return contacts, nil
}

// FindContactsRequestByUserIDAndFriendID 查询是否已经发了好友请求
func FindContactsRequestByUserIDAndFriendID(userid, friendid uint64, cols ...string) (*ContactsRequest, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	contactsRequest := &ContactsRequest{}
	tx := database.Scopes(UseContactsRequestTable(userid))
	if len(cols) > 0 {
		tx.Select(cols)
	}
	tx.Where("user_id = ? And friend_id = ?", userid, friendid).Last(contactsRequest)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}

		return nil, tx.Error
	}

	return contactsRequest, nil
}

// DeleteContactsRequestByUserIDAndFriendID 删除
func DeleteContactsRequestByUserIDAndFriendID(userid, friendid uint64) error {
	if database == nil {
		return gorm.ErrInvalidDB
	}

	tx := database.Scopes(UseContactsRequestTable(userid)).Where("user_id = ? And friend_id = ?", userid, friendid).Delete(&ContactsRequest{})
	return tx.Error
}

// CreateContactsRequest 创建
func CreateContactsRequest(contactsRequest *ContactsRequest) error {
	if database == nil {
		return gorm.ErrInvalidDB
	}

	tx := database.Scopes(UseContactsRequestTable(contactsRequest.UserID)).Create(contactsRequest)
	return tx.Error
}

// AddContactsFromRequest 添加联系人
func AddContactsFromRequest(users, friend *Users, request *ContactsRequest, remark string) error {
	if database == nil {
		return gorm.ErrInvalidDB
	}

	topic := helper.GenerateTopic(users.ID, friend.ID)
	contactsA := &Contacts{
		UserID:    users.ID,
		FriendID:  friend.ID,
		FNickname: friend.Nickname,
		FHeadimg:  friend.Headimg,
		FSex:      friend.Sex,
		Type:      ContactsTypePerson,
		Topic:     topic,
		Remark:    remark,
		Status:    0,
		Version:   time.Now().Unix(),
	}

	contactsB := &Contacts{
		UserID:    friend.ID,
		FriendID:  users.ID,
		FNickname: users.Nickname,
		FHeadimg:  users.Headimg,
		FSex:      users.Sex,
		Type:      ContactsTypePerson,
		Topic:     topic,
		Remark:    request.Remark,
		Status:    0,
		Version:   time.Now().Unix(),
	}

	return database.Transaction(func(tx *gorm.DB) error {
		contacts := &Contacts{}
		contactsRequest := &ContactsRequest{}
		ret := tx.Scopes(UseContactsTable(contactsA.UserID)).Where("user_id = ? And friend_id = ?", contactsA.UserID, contactsA.FriendID).Take(contacts)
		if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
			ret = tx.Scopes(UseContactsTable(contactsA.UserID)).Create(contactsA)
			if ret.Error != nil {
				return ret.Error
			}
		}

		ret = tx.Scopes(UseContactsTable(contactsB.UserID)).Where("user_id = ? And friend_id = ?", contactsB.UserID, contactsB.FriendID).Take(contacts)
		if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
			ret = tx.Scopes(UseContactsTable(contactsB.UserID)).Create(contactsB)
			if ret.Error != nil {
				return ret.Error
			}
		}

		ret = tx.Scopes(UseContactsRequestTable(request.UserID)).Where("user_id = ? And friend_id = ?", contactsA.UserID, contactsA.FriendID).Delete(contactsRequest)
		if ret.Error != nil {
			return ret.Error
		}

		tx.Scopes(UseContactsRequestTable(request.UserID)).Where("user_id = ? And friend_id = ?", contactsB.UserID, contactsB.FriendID).Delete(contactsRequest)
		return nil
	})
}

// RefuseAddContact 拒绝
func RefuseAddContact(users, friend *Users, request *ContactsRequest) error {
	if database == nil {
		return gorm.ErrInvalidDB
	}

	contactsRequest := &ContactsRequest{
		UserID:    friend.ID,
		FriendID:  users.ID,
		FromID:    request.FromID,
		FNickname: users.Nickname,
		FHeadimg:  users.Headimg,
		FSex:      users.Sex,
		Remark:    "",
		Status:    1,
		Messages:  request.Messages,
		Version:   time.Now().Unix(),
	}

	return database.Transaction(func(tx *gorm.DB) error {
		ret := tx.Scopes(UseContactsRequestTable(users.ID)).Where("user_id = ? And friend_id = ?", users.ID, friend.ID).Updates(map[string]interface{}{"status": 1, "messages": request.Messages, "version": time.Now().Unix()})
		if ret.Error != nil {
			return ret.Error
		}

		tx.Scopes(UseContactsRequestTable(friend.ID)).Where("user_id = ? And friend_id = ?", friend.ID, users.ID).Delete(&ContactsRequest{})
		ret = tx.Scopes(UseContactsRequestTable(friend.ID)).Create(contactsRequest)
		if ret.Error != nil {
			return ret.Error
		}
		return nil
	})
}

// UpdateContactsRequestStatusByID 更新状态
func UpdateContactsRequestStatusByID(userid, friendid uint64, status int) error {
	if database == nil {
		return gorm.ErrInvalidDB
	}

	tx := database.Scopes(UseContactsRequestTable(userid)).Where("user_id = ? And friend_id = ?", userid, friendid).Updates(map[string]interface{}{"status": status, "version": time.Now().Unix()})
	return tx.Error
}

// FindContactsRequestByUserID 获取列表
func FindContactsRequestByUserID(userid uint64, page, size int32, version int64, cols ...string) ([]*ContactsRequest, int64, error) {
	if page < 1 {
		page = 1
	}

	if size > 50 {
		size = 50
	} else if size < 1 {
		size = 10
	}

	offset := (page - 1) * size
	if database == nil {
		return nil, 0, gorm.ErrInvalidDB
	}

	var count int64
	data := make([]*ContactsRequest, 0)

	tx := database.Scopes(UseContactsRequestTable(userid)).Select("id").Where("user_id = ? AND version > ? ", userid, version).Count(&count)
	if tx.Error != nil {
		return nil, 0, gorm.ErrInvalidDB
	}

	tx = database.Scopes(UseContactsRequestTable(userid))
	if len(cols) > 0 {
		tx.Select(cols)
	}

	tx.Where("user_id = ? AND version > ? ", userid, version).Order("id DESC").Offset(int(offset)).Limit(int(size)).Find(&data)
	if tx.Error != nil {
		return nil, 0, gorm.ErrInvalidDB
	}

	return data, count, nil
}

// FindContactsByUserIDPage 获取联系人
func FindContactsByUserIDPage(userid uint64, page, size int32, version int64, cols ...string) ([]*Contacts, int64, error) {
	if page < 1 {
		page = 1
	}

	if size > 50 {
		size = 50
	} else if size < 1 {
		size = 10
	}

	offset := (page - 1) * size
	if database == nil {
		return nil, 0, gorm.ErrInvalidDB
	}

	var count int64
	data := make([]*Contacts, 0)

	tx := database.Scopes(UseContactsTable(userid)).Where("user_id = ? AND version > ? AND status = 0", userid, version).Count(&count)
	if tx.Error != nil {
		return nil, 0, gorm.ErrInvalidDB
	}

	tx = database.Scopes(UseContactsTable(userid))
	if len(cols) > 1 {
		tx.Select(cols)
	}
	tx.Where("user_id = ? AND version > ? AND status = 0", userid, version).Offset(int(offset)).Limit(int(size)).Find(&data)
	if tx.Error != nil {
		return nil, 0, gorm.ErrInvalidDB
	}

	return data, count, nil
}

// FindContactsByUserID 获取联系人
func FindContactsByUserID(userid uint64, members []uint64, cols ...string) ([]*Contacts, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	data := make([]*Contacts, 0)
	tx := database.Scopes(UseContactsTable(userid))
	if len(cols) > 1 {
		tx.Select(cols)
	}
	tx.Where("user_id = ? AND friend_id IN ? AND status = 0", userid, members).Find(&data)
	if tx.Error != nil {
		return nil, gorm.ErrInvalidDB
	}

	return data, nil
}

func CreateContacts(contact *Contacts) error {
	if database == nil {
		return gorm.ErrInvalidDB
	}

	tx := database.Scopes(UseContactsTable(contact.UserID)).Create(contact)
	return tx.Error
}
