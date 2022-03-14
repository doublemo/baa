package dao

import (
	"strconv"
	"time"

	"gorm.io/gorm"
)

type (
	// RobotContacts 通讯录
	RobotContacts struct {
		ID          uint64 `gorm:"<-:create;primaryKey;autoIncrement"`
		UserID      uint64 `gorm:"index;index:userid_friendid;priority:1"`
		FriendID    uint64 `gorm:"index:userid_friendid;priority:2"`
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
)

func CreateOrUpdateRobotContacts(contacts ...*RobotContacts) error {
	if database == nil {
		return gorm.ErrInvalidDB
	}

	idMap := make(map[uint64][]uint64)
	for _, contact := range contacts {
		if _, ok := idMap[contact.UserID]; !ok {
			idMap[contact.UserID] = make([]uint64, 0)
		}
		idMap[contact.UserID] = append(idMap[contact.UserID], contact.FriendID)
	}

	dataMap := make(map[string]*RobotContacts)
	for k, v := range idMap {
		robotContacts := make([]*RobotContacts, 0)
		tx := database.Where("user_id = ? AND friend_id IN ?", k, v).Find(&robotContacts)
		if tx.Error != nil {
			return tx.Error
		}
		for _, r := range robotContacts {
			dataMap[strconv.FormatUint(r.UserID, 10)+strconv.FormatUint(r.FriendID, 10)] = r
		}
	}

	createData := make([]*RobotContacts, 0)
	updateData := make([]*RobotContacts, 0)
	for _, r := range contacts {
		if m, ok := dataMap[strconv.FormatUint(r.UserID, 10)+strconv.FormatUint(r.FriendID, 10)]; ok && m != nil {
			r.ID = m.ID
			updateData = append(updateData, r)
		} else {
			r.CreateAt = time.Now().Unix()
			createData = append(createData, r)
		}
	}

	if len(createData) > 0 {
		tx := database.Create(createData)
		if tx.Error != nil {
			return tx.Error
		}
	}

	if len(updateData) > 0 {
		for _, r := range updateData {
			if tx := database.Save(r); tx.Error != nil {
				return tx.Error
			}
		}
	}
	return nil
}
