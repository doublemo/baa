package dao

// Friends 好友
type Friends struct {
	ID          uint64
	UserID      uint64
	FriendID    uint64
	FNickname   string
	FHeadimg    string
	FSex        int8
	Remark      string
	Mute        int8 // 消息免打扰
	StickyOnTop int8 // 聊天置顶
}
