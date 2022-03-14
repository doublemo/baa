package dao

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RobotVersionManagers 数据版本管理
type RobotVersionManagers struct {
	ID        string `gorm:"size:255;primaryKey"`
	UserID    uint64 `gorm:"size:255;index"`
	VersionID int64  `gorm:"size:255"`
	Version   string
}

func FindVersionByID(id string, userid uint64) (*RobotVersionManagers, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	v := &RobotVersionManagers{}
	tx := database.Where("id = ? AND user_id = ?", id, userid).First(v)
	return v, tx.Error
}

func FindVersionByUserID(userid uint64) (map[string]*RobotVersionManagers, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	v := make([]*RobotVersionManagers, 0)
	tx := database.Where("user_id = ?", userid).Find(&v)
	if tx.Error != nil {
		return nil, tx.Error
	}

	data := make(map[string]*RobotVersionManagers)
	for _, value := range v {
		data[value.ID] = value
	}
	return data, nil
}

func UpsertVersionByID(v *RobotVersionManagers) error {
	if database == nil {
		return gorm.ErrInvalidDB
	}

	tx := database.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"version_id", "version"}),
	}).Create(v)
	return tx.Error
}
