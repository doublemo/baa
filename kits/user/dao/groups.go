package dao

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

const (
	defaultGroupsMaxRecord       = 10000000
	defaultGroupsMaxTable        = 50
	defaultGroupMembersMaxRecord = 10000000
	defaultGroupMembersMaxTable  = 100
)

const (
	GroupMemberOfficialTitleNone  = 0  // 普通成员
	GroupMemberOfficialTitleOwner = 10 // 群主创建者
)

const (
	GroupMembersStatusInvitationNotSent = -2 // 还未发出邀请
	GroupMembersStatusWaitingJoin       = -1 // 等待加入的用户
	GroupMembersStatusNormal            = 0  // 正常的用户
)

const (
	GroupMembersOriginVolunteer = 0 // 自己加入
	GroupMembersOriginInvite    = 1 // 通过邀请加入
)

type (
	// Groups 群组
	Groups struct {
		ID      uint64 `gorm:"<-:create;primaryKey"`
		Name    string `gorm:"size:50"`
		Notice  string `gorm:"size:1000"`
		Headimg string `gorm:"size:255"`
		UserID  uint64 `gorm:"index"`
		Topic   uint64
	}

	// GroupMembers 群用户
	GroupMembers struct {
		ID            uint64 `gorm:"<-:create;primaryKey;autoIncrement"`
		GroupID       uint64 `gorm:"<-:create;index;index:group_id_user_id"`
		UserID        uint64 `gorm:"<-:create;index:group_id_user_id"`
		Nickname      string `gorm:"size:50"`
		Headimg       string `gorm:"size:255"`
		Sex           int8
		Remark        string `gorm:"size:50"`
		Mute          int8   // 消息免打扰
		StickyOnTop   int8   // 聊天置顶
		Alias         string `gorm:"size:50"`
		Topic         uint64
		OfficialTitle int8
		Status        int8
		CreateAt      int64  `gorm:"autoCreateTime"`
		JoinAt        int64  // 加入时间
		Origin        int8   // 来源 0 自己加入  1 邀请加入
		Handler       uint64 // 办理用户加入的人ID
		Version       int64
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
func UseGroupMembersTable(id uint64) func(tx *gorm.DB) *gorm.DB {
	return useTable(id, &GroupMembers{}, defaultGroupMembersMaxRecord, defaultGroupMembersMaxTable)
}

// FindGroupsMembersByGroupID 获取群成员信息
func FindGroupsMembersByGroupID(id uint64, page, size int32, version int64, cols ...string) ([]*GroupMembers, int64, error) {
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
	data := make([]*GroupMembers, 0)

	tx := database.Scopes(UseGroupMembersTable(id)).Where("group_id = ? AND version > ?", id, version).Count(&count)
	if tx.Error != nil {
		return nil, 0, gorm.ErrInvalidDB
	}

	tx = database.Scopes(UseGroupMembersTable(id))
	if len(cols) > 1 {
		tx.Select(cols)
	}
	tx.Where("group_id = ? AND version > ?", id, version).Offset(int(offset)).Limit(int(size)).Find(&data)
	if tx.Error != nil {
		return nil, 0, gorm.ErrInvalidDB
	}

	return data, count, nil
}

// FindGroupsByGroupID 获取群信息
func FindGroupsByGroupID(id uint64, cols ...string) (*Groups, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	groups := &Groups{}
	tx := database.Scopes(UseGroupsTable(id))
	if len(cols) > 0 {
		tx.Select(cols)
	}
	tx.Where("id = ?", id).Last(groups)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}

		return nil, tx.Error
	}

	return groups, nil
}

// FindGroupsMembersIDByGroupID 获取群成员ID
func FindGroupsMembersIDByGroupID(id uint64, page, size int32) ([]uint64, int64, error) {
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
	data := make([]uint64, 0)

	tx := database.Scopes(UseGroupMembersTable(id)).Where("group_id = ?", id).Count(&count)
	if tx.Error != nil {
		return nil, 0, gorm.ErrInvalidDB
	}

	tx = database.Scopes(UseGroupMembersTable(id))
	tx.Where("group_id = ?", id).Offset(int(offset)).Limit(int(size)).Pluck("user_id", &data)
	if tx.Error != nil {
		return nil, 0, gorm.ErrInvalidDB
	}
	return data, count, nil
}

func FindGroupsMemberByGroupIDAndUserID(groupid, userid uint64, cols ...string) (*GroupMembers, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	groupMembers := &GroupMembers{}
	tx := database.Scopes(UseGroupMembersTable(groupid))
	if len(cols) > 0 {
		tx.Select(cols)
	}
	tx.Where("group_id = ? AND user_id = ?", groupid, userid).Last(groupMembers)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}

		return nil, tx.Error
	}

	return groupMembers, nil
}

func FindGroupsMembersByGroupIDAndUserID(groupid uint64, userid []uint64, cols ...string) ([]*GroupMembers, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	groupMembers := make([]*GroupMembers, 0)
	tx := database.Scopes(UseGroupMembersTable(groupid))
	if len(cols) > 0 {
		tx.Select(cols)
	}
	tx.Where("group_id = ? AND user_id IN ?", groupid, userid).Find(&groupMembers)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}

		return nil, tx.Error
	}

	return groupMembers, nil
}

func FindGroupsMembersNotNormalByGroupIDAndUserID(groupid uint64, userid []uint64, cols ...string) ([]*GroupMembers, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	groupMembers := make([]*GroupMembers, 0)
	tx := database.Scopes(UseGroupMembersTable(groupid))
	if len(cols) > 0 {
		tx.Select(cols)
	}
	tx.Where("group_id = ? AND user_id IN ? AND status < ?", groupid, userid, GroupMembersStatusNormal).Find(&groupMembers)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}

		return nil, tx.Error
	}

	return groupMembers, nil
}

func CreateAndJoinGroup(group *Groups, members ...*GroupMembers) error {
	if database == nil {
		return gorm.ErrInvalidDB
	}

	return database.Transaction(func(tx *gorm.DB) error {
		r := tx.Scopes(UseGroupsTable(group.ID)).Create(group)
		if r.Error != nil {
			return r.Error
		}

		if r.RowsAffected < 1 {
			return errors.New("CreateFailed")
		}

		r = tx.Scopes(UseContactsTable(group.UserID)).Create(&Contacts{
			UserID:    group.UserID,
			FriendID:  group.ID,
			FNickname: group.Name,
			FHeadimg:  group.Headimg,
			FSex:      0,
			Type:      ContactsTypeGroup,
			Topic:     group.Topic,
			Status:    0,
			Version:   time.Now().Unix(),
		})

		if r.Error != nil {
			return r.Error
		}

		if r.RowsAffected < 1 {
			return errors.New("CreateFailed")
		}

		r = tx.Scopes(UseGroupMembersTable(group.ID)).Create(members)
		if r.Error != nil {
			return r.Error
		}

		if r.RowsAffected < 1 {
			return errors.New("CreateFailed")
		}
		return nil
	})
}

func UpdateGroupMembersStatus(id uint64, status int, members ...uint64) error {
	if database == nil {
		return gorm.ErrInvalidDB
	}

	tx := database.Scopes(UseGroupMembersTable(id)).Where("group_id = ? AND user_id IN ?", id, members).Update("status", status)
	return tx.Error
}

func CreateGroupsMember(id uint64, members ...*GroupMembers) error {
	if database == nil {
		return gorm.ErrInvalidDB
	}

	return database.Transaction(func(tx *gorm.DB) error {
		r := tx.Scopes(UseGroupMembersTable(id)).Create(members)
		if r.Error != nil {
			return r.Error
		}

		if r.RowsAffected < 1 {
			return errors.New("CreateFailed")
		}
		return nil
	})
}
