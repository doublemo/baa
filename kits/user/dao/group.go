package dao

type (
	// Groups 群组
	Groups struct {
		ID      uint64
		Name    string
		Notice  string
		Headimg string
	}

	// GroupUsers 群用户
	GroupUsers struct {
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
	}
)
