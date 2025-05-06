//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package system

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/sirupsen/logrus"
)

func FindProcessByPidFile(f string) (*process.Process, error) {
	b, err := os.ReadFile(f)
	if errors.Is(err, os.ErrNotExist) {
		return nil, process.ErrorProcessNotRunning
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read pid file: %w", err)
	}

	pidStr := strings.TrimSpace(string(b))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PID format: %w", err)
	}

	logrus.Infof("pid file %q contains pid %d", f, pid)

	return FindProcessByPid(int32(pid))
}

// KillExpectProcNameFromPPIDFile It only kills the process if it matches the expected name
func KillExpectProcNameFromPPIDFile(f, expectedName string) error {
	proc, err := FindProcessByPidFile(f)
	// if pid not running, do nothing
	if errors.Is(err, process.ErrorProcessNotRunning) {
		logrus.Infof("process not running, do nothing")
		return nil
	}
	// other errors must be returned
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	procName, err := proc.Name()
	if err != nil {
		return fmt.Errorf("failed to get process name: %w", err)
	}

	// if pid is running but the pid of process name is not name, do nothing
	if procName != expectedName {
		return nil
	}

	logrus.Infof("process name is %q, kill it", procName)
	return proc.Kill() //nolint:wrapcheck
}

func FindProcessByPid(pid int32) (*process.Process, error) {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("failed to find process id %d: %w", pid, err)
	}

	return proc, nil
}
