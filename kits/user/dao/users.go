package dao

// Users 用户表
type Users struct {
	ID       uint64
	Nickname string
	Headimg  string
	Age      int8
	Sex      int8
	Idcard   string
}
