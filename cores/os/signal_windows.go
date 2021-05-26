// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

package os

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// handleSignals 在windows系统中运行.
// 处理系统信息
func handleSignals(s Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for sig := range c {
			log.Printf("Trapped %q signal\n", sig)
			s.Shutdown()
			os.Exit(0)
		}
	}()
}

// ProcessSignal 向系统正运行的进程发送系统信号
func ProcessSignal(command Command, process string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}

	defer m.Disconnect()

	s, err := m.OpenService(process)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}

	defer s.Close()

	var (
		cmd svc.Cmd
		to  svc.State
	)

	switch command {
	case CommandQuit, CommandStop:
		cmd = svc.Stop
		to = svc.Stopped

	case CommandReload:
		cmd = svc.ParamChange
		to = svc.Running

	case CommandUSR1:
		cmd = windowsUsrCode1
		to = svc.Running

	case CommandUSR2:
		cmd = windowsUsrCode2
		to = svc.Running
	default:
		return fmt.Errorf("unknown signal %q", command)
	}

	status, err := s.Control(cmd)
	if err != nil {
		return fmt.Errorf("could not send control=%d: %v", cmd, err)
	}

	timeout := time.Now().Add(10 * time.Second)
	for status.State != to {
		if timeout.Before(time.Now()) {
			return fmt.Errorf("timeout waiting for service to go to state=%d", to)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
	}

	return nil
}
