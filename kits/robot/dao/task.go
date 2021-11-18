package dao

import "gorm.io/gorm"

type (
	// Tasks 任务表
	Tasks struct {
		ID          uint64 `gorm:"<-:create;primaryKey;autoIncrement"`
		Name        string `gorm:"size:255"`
		ScriptPath  string `gorm:"size:255"`
		Description string `gorm:"size:5000"`
		Type        int    // 任务类型 0 单次任务 1 多次任务 2 定时任务 3 常住任务
		Parameters  string `gorm:"size:2000"`
		RunMode     int    // 运行方式 0 同步 1 异步
	}

	// RobotsTasks 机器人任务
	RobotsTasks struct {
		ID         uint64 `gorm:"<-:create;primaryKey;autoIncrement"`
		TaskID     uint64 `gorm:"<-:create;"`
		RobotID    uint64 `gorm:"<-:create;index;index:robot_id_task_group"`
		Parameters string `gorm:"size:2000"`
		TaskGroup  int    `gorm:"<-:create;index:robot_id_task_group"` // 任务分组
		Ordered    int    // 任务顺序
		CreateAt   int64  `gorm:"autoCreateTime"`
	}
)

// FindTasksByID 获取任务
func FindTasksByID(id uint64) (*Tasks, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	tasks := &Tasks{ID: id}
	tx := database.First(tasks)
	return tasks, tx.Error
}

// FindTasksInID 获取任务
func FindTasksInID(values []uint64, cols ...string) ([]*Tasks, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	tasks := make([]*Tasks, 0)
	tx := database.Where("id IN ?", values)
	if len(cols) > 0 {
		tx.Select(cols)
	}

	tx.Find(&tasks)
	return tasks, tx.Error
}

// FindTasks 获取任务
func FindTasks(cols ...string) ([]*Tasks, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	tasks := make([]*Tasks, 0)
	var tx *gorm.DB
	if len(cols) > 0 {
		tx = database.Select(cols).Find(&tasks)
	} else {
		tx = database.Find(&tasks)
	}
	return tasks, tx.Error
}

// FindRobotsTasksByRobotIDAndGroup 获取机器人任务
func FindRobotsTasksByRobotIDAndGroup(id uint64, group int, cols ...string) ([]*RobotsTasks, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	tasks := make([]*RobotsTasks, 0)
	tx := database.Where("robot_id = ? AND task_group = ?", id, group).Order("ordered ASC")
	if len(cols) > 0 {
		tx.Select(cols)
	}

	tx.Find(&tasks)
	return tasks, tx.Error
}
