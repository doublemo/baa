package dao

import "gorm.io/gorm"

type (
	// Robots 机器人账户存储表
	Robots struct {
		ID        uint64 `gorm:"<-:create;primaryKey;autoIncrement"`
		AccountID uint64 `gorm:"<-:create;index"`
		UnionID   uint64 `gorm:"<-:create;index"`
		UserID    uint64 `gorm:"<-:create;index"`
		Schema    string `gorm:"<-:create;size:50;index:schema_name"`
		Name      string `gorm:"<-:create;index:schema_name"`
		Secret    string `gorm:"size:255"`
		IndexNo   string `gorm:"size:50"`
		Nickname  string `gorm:"size:50"`
		Headimg   string `gorm:"size:255"`
		Age       int8
		Sex       int8
		Idcard    string `gorm:"size:50"`
		Phone     string `gorm:"size:20"`
		CreatedAt int64  `gorm:"autoCreateTime"`
	}

	// RobotsSettings 机器人设置
	RobotsSettings struct {
		ID uint64 `gorm:"<-:create;primaryKey"`
	}
)

// CreateRobot 创建机器人
func CreateRobot(robot *Robots) error {
	if database == nil {
		return gorm.ErrInvalidDB
	}

	return database.Create(robot).Error
}
