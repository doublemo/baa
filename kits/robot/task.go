package robot

import (
	"errors"

	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/kits/robot/dao"
	"github.com/doublemo/baa/kits/robot/session"
)

func execTask(peer session.Peer, task *pb.Robot_Start_Robot, c RobotConfig) error {
	robotsTasks, err := dao.FindRobotsTasksByRobotIDAndGroup(task.ID, int(task.TaskGroup))
	if err != nil {
		return err
	}

	if len(robotsTasks) < 1 {
		return errors.New("No task can be run")
	}

	tasksIds := make([]uint64, len(robotsTasks))
	for k, v := range robotsTasks {
		tasksIds[k] = v.TaskID
	}

	tasks, err := dao.FindTasksInID(tasksIds)
	if err != nil {
		return err
	}

	tasksMap := make(map[uint64]*dao.Tasks)
	for _, v := range tasks {
		tasksMap[v.ID] = v
	}

	for _, rt := range robotsTasks {
		if _, ok := tasksMap[rt.TaskID]; ok {
			continue
		}
	}
	return nil
}

// 一次性任务
// 循环任务
// 定时任务
