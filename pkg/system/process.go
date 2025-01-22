//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package system

import (
	"fmt"
	"os/exec"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/sirupsen/logrus"
)

func KillCmdWithWarn(cmd ...*exec.Cmd) {
	for _, cmd := range cmd {
		if cmd != nil {
			logrus.Warnf("Killing process PID: %d, PATH: %q", cmd.Process.Pid, cmd.Path)
			_ = cmd.Process.Kill()
		}
	}
}

func GetPPID(pid int32) (int32, error) {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return -1, fmt.Errorf("failed to get process %d: %w", pid, err)
	}
	ppid, err := proc.Ppid()
	if err != nil {
		return -1, fmt.Errorf("failed to get parent process id for %d: %w", pid, err)
	}
	return ppid, nil
}

// FindProcessByPath find process by path, return *process.Process, if it has error return error
func FindProcessByPath(path string) (*process.Process, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %w", err)
	}

	for _, proc := range procs {
		exe, err := proc.Exe()
		if err != nil {
			continue
		}
		if exe == path {
			return proc, nil
		}
	}
	return nil, fmt.Errorf("process with path %s not found", path)
}
