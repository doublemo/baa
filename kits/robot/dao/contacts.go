package dao

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
)
