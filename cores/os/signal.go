// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

// +build !windows

package os

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

// handleSignals 在除windows以外的系统中运行.
// 处理系统信息
func handleSignals(s Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGHUP)
	go func() {
		for {
			select {
			case sig := <-c:
				switch sig {
				case syscall.SIGINT:
					s.Shutdown()
					os.Exit(0)

				case syscall.SIGUSR1:
					s.OtherCommand(1)

				case syscall.SIGUSR2:
					s.OtherCommand(2)

				case syscall.SIGHUP:
					// reload
					s.Reload()
				}

			case <-s.QuitCh():
				return
			}
		}
	}()
}

// ProcessSignal 向系统正运行的进程发送系统信号
func ProcessSignal(command Command, process string) error {
	var pid int
	re := regexp.MustCompile("^([1-9])\\d+")
	if !re.MatchString(process) {
		pids, err := resolvePids(process)
		if err != nil {
			return err
		}

		if len(pids) == 0 {
			return fmt.Errorf("no %s processes running", process)
		}

		if len(pids) > 1 {
			errStr := fmt.Sprintf("multiple %s processes running:\n", process)
			prefix := ""
			for _, p := range pids {
				errStr += fmt.Sprintf("%s%d", prefix, p)
				prefix = "\n"
			}
			return errors.New(errStr)
		}

		pid = pids[0]
	} else {
		p, err := strconv.Atoi(process)
		if err != nil {
			return fmt.Errorf("invalid pid: %s", process)
		}
		pid = p
	}

	var err error
	switch command {
	case CommandStop:
		err = kill(pid, syscall.SIGKILL)
	case CommandQuit:
		err = kill(pid, syscall.SIGINT)
	case CommandUSR1:
		err = kill(pid, syscall.SIGUSR1)
	case CommandReload:
		err = kill(pid, syscall.SIGHUP)
	case CommandUSR2:
		err = kill(pid, syscall.SIGUSR2)
	default:
		err = fmt.Errorf("unknown signal %q", command)
	}
	return err
}

// resolvePids 查询返回所有在系统中指定的进程
func resolvePids(name string) ([]int, error) {
	// If pgrep isn't available, this will just bail out and the user will be
	// required to specify a pid.
	output, err := pgrep(name)
	if err != nil {
		switch err.(type) {
		case *exec.ExitError:
			// ExitError indicates non-zero exit code, meaning no processes
			// found.
			break
		default:
			return nil, errors.New("unable to resolve pid, try providing one")
		}
	}
	var (
		myPid   = os.Getpid()
		pidStrs = strings.Split(string(output), "\n")
		pids    = make([]int, 0, len(pidStrs))
	)
	for _, pidStr := range pidStrs {
		if pidStr == "" {
			continue
		}
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			return nil, errors.New("unable to resolve pid, try providing one")
		}
		// Ignore the current process.
		if pid == myPid {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

func kill(pid int, signal syscall.Signal) error {
	return syscall.Kill(pid, signal)
}

func pgrep(name string) ([]byte, error) {
	return exec.Command("pgrep", name).Output()
}
