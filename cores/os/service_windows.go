// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

package os

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	windowsUsrCode1 = 128
	windowsUsrCMD1  = svc.Cmd(windowsUsrCode1)
	windowsUsrCode2 = 129
	windowsUsrCMD2  = svc.Cmd(windowsUsrCode2)
	acceptUsr       = svc.Accepted(windowsUsrCMD1)
)

// ServiceWindowsConfig windows服务配置信息
type ServiceWindowsConfig struct {
	// name windows 服务名称
	Name string

	// DisplayName 设置在windows服务信息中显示的名称
	DisplayName string

	// Description 服务信息的描述
	Description string

	// UserName windows 启动这个服务所需要的用户
	UserName string

	// Password windows 启动这个服务用户的密码
	Password string

	// Dependencies 依赖关系信息
	Dependencies []string

	// Arguments 启动这个服务所需要的参数
	Arguments []string
}

// serviceWindows 服务
type serviceWindows struct {
	s Server
}

// Execute 用于windows运行服务
func (w *serviceWindows) Execute(args []string, changes <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	status <- svc.Status{State: svc.StartPending}
	go func() {
		w.s.Start()
	}()

	if !w.s.Readyed() {
		return false, 1
	}

	status <- svc.Status{
		State:   svc.Running,
		Accepts: svc.AcceptStop | svc.AcceptShutdown | svc.AcceptParamChange | acceptUsr,
	}

loop:
	for change := range changes {
		switch change.Cmd {
		case svc.Interrogate:
			status <- change.CurrentStatus
		case svc.Stop, svc.Shutdown:
			w.s.Shutdown()
			break loop

		case windowsUsrCMD1:
			w.s.OtherCommand(1)

		case windowsUsrCMD2:
			w.s.OtherCommand(2)

		case svc.ParamChange:
			w.s.Reload()

		default:
			log.Printf("Unexpected control request: %v \n", change.Cmd)
		}
	}

	status <- svc.Status{State: svc.StopPending}
	return false, 0
}

// runService 运行服务
func runService(s Server) error {
	isInteractive, err := svc.IsAnInteractiveSession()
	if err != nil {
		return err
	}

	if isInteractive {
		return s.Start()
	}
	return svc.Run(s.ServiceName(), &serviceWindows{s: s})
}

// IsWindowsService 确认是否已windows服务方式进行运行
func IsWindowsService() bool {
	isInteractive, _ := svc.IsAnInteractiveSession()
	return !isInteractive
}

// InstallWindowsService 安装windows服务
func InstallWindowsService(c *ServiceWindowsConfig) error {
	exepath, err := ExecPath()
	if err != nil {
		return err
	}

	m, err := mgr.Connect()
	if err != nil {
		return err
	}

	defer m.Disconnect()
	s, err := m.OpenService(c.Name)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", c.Name)
	}

	s, err = m.CreateService(c.Name, exepath, mgr.Config{
		DisplayName:      c.DisplayName,
		Description:      c.Description,
		StartType:        mgr.StartAutomatic,
		ServiceStartName: c.UserName,
		Password:         c.Password,
		Dependencies:     c.Dependencies,
	}, c.Arguments...)

	if err != nil {
		return err
	}

	defer s.Close()
	err = eventlog.InstallAsEventCreate(c.Name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		if !strings.Contains(err.Error(), "exists") {
			s.Delete()
			return fmt.Errorf("SetupEventLogSource() failed: %s", err)
		}
	}

	return nil
}

// UninstallWindowsService 卸载windows服务
func UninstallWindowsService(serviceName string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s is not installed", serviceName)
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(serviceName)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}
	return nil
}

// StatusWindowsService 获取windows服务状态
func StatusWindowsService(serviceName string) (svc.State, error) {
	m, err := mgr.Connect()
	if err != nil {
		return svc.Stopped, err
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return svc.Stopped, err
	}

	status, err := s.Query()
	if err != nil {
		return svc.Stopped, err
	}

	return status.State, nil
}

// StartWindowsService 启动windows服务
func StartWindowsService(serviceName string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return err
	}
	defer s.Close()
	return s.Start()
}

// StopWindowsService 停止windows服务
func StopWindowsService(serviceName string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return err
	}
	defer s.Close()
	return stopWaitWindowsService(s)
}

// RestartWindowsService 重启windows服务
func RestartWindowsService(serviceName string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return err
	}
	defer s.Close()

	err = stopWaitWindowsService(s)
	if err != nil {
		return err
	}

	return s.Start()
}

// stopWaitWindowsService 等待windows服务关闭
func stopWaitWindowsService(s *mgr.Service) error {
	status, err := s.Control(svc.Stop)
	if err != nil {
		return err
	}

	timeDuration := time.Millisecond * 50
	timeout := time.After(getStopTimeout() + (timeDuration * 2))
	tick := time.NewTicker(timeDuration)
	defer tick.Stop()

	for status.State != svc.Stopped {
		select {
		case <-tick.C:
			status, err = s.Query()
			if err != nil {
				return err
			}
		case <-timeout:
			break
		}
	}
	return nil
}

// getStopTimeout fetches the time before windows will kill the service.
func getStopTimeout() time.Duration {
	// For default and paths see https://support.microsoft.com/en-us/kb/146092
	defaultTimeout := time.Millisecond * 20000
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control`, registry.READ)
	if err != nil {
		return defaultTimeout
	}
	sv, _, err := key.GetStringValue("WaitToKillServiceTimeout")
	if err != nil {
		return defaultTimeout
	}
	v, err := strconv.Atoi(sv)
	if err != nil {
		return defaultTimeout
	}
	return time.Millisecond * time.Duration(v)
}
