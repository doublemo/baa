package command

import (
	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/internal/proto/kit"
)

const (
	// UserContacts 通讯录操作
	UserContacts coresproto.Command = kit.User + (iota + 1)

	// UserContactsList 联系人列表
	UserContactsList

	// UserContactsRequest 请求添加好友列表
	UserContactsRequest

	// UserRegister 注册
	UserRegister

	// UserCheckIsMyFriend 检查是否是好朋友
	UserCheckIsMyFriend

	// UserCheckInGroup 检查是否在群中
	UserCheckInGroup

	// UserGroupMembers  获取群成员信息
	UserGroupMembers

	// UserGroupMembersValidID  获取群有效成员ID
	UserGroupMembersValidID
)
