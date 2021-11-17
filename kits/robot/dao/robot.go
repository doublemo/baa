package dao

import "gorm.io/gorm"

type (
	// Robots 机器人账户存储表
	Robots struct {
		ID         uint64 `gorm:"<-:create;primaryKey;autoIncrement"`
		AccountID  uint64 `gorm:"<-:create;index"`
		UnionID    uint64 `gorm:"<-:create;index"`
		UserID     uint64 `gorm:"<-:create;index"`
		SchemaName string `gorm:"<-:create;size:50;index:schema_name"`
		Name       string `gorm:"<-:create;index:schema_name"`
		Secret     string `gorm:"size:255"`
		IndexNo    string `gorm:"size:50"`
		Nickname   string `gorm:"size:50"`
		Headimg    string `gorm:"size:255"`
		Age        int8
		Sex        int8
		Idcard     string `gorm:"size:50"`
		Phone      string `gorm:"size:20"`
		Agent      string `gorm:"size:50"`
		CreatedAt  int64  `gorm:"autoCreateTime"`
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

// UpdatesRobotsByID 更新数据
func UpdatesRobotsByID(id uint64, col string, value interface{}) (int64, error) {
	if database == nil {
		return 0, gorm.ErrInvalidDB
	}

	r := database.Model(&Robots{}).Where("id = ?", id).Update(col, value)
	return r.RowsAffected, r.Error
}

// FindRobotsByAccountID 查询机器人
func FindRobotsByAccountID(id uint64, cols ...string) (*Robots, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	robots := &Robots{}
	tx := database.Where("account_id = ?", id)
	if len(cols) > 0 {
		tx.Select(cols)
	}

	tx.First(robots)
	return robots, tx.Error
}

// FindRobotsBySchemaName 查询机器人
func FindRobotsBySchemaName(schema, name string, cols ...string) (*Robots, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	robots := &Robots{}
	tx := database.Where("schema_name = ? AND name = ?", schema, name)
	if len(cols) > 0 {
		tx.Select(cols)
	}

	tx.First(robots)
	return robots, tx.Error
}

// FindRobotsInID 查询机器人
func FindRobotsInID(id ...uint64) ([]*Robots, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	robots := make([]*Robots, 0)
	tx := database.Where("id IN ?", id)
	tx.Select("id", "account_id", "union_id", "user_id", "schema_name", "name", "secret", "index_no", "nickname", "headimg", "sex", "created_at")
	tx.Find(&robots)
	return robots, tx.Error
}
