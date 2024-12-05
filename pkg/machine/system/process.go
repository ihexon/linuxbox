//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package system

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/process"
)

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

func IsProcesSAlive(pids []int32) (bool, error) {
	var (
		isRunning = false
		targetPid int32
		err       error
	)

	for _, pid := range pids {
		targetPid = pid
		isRunning, err = IsProcessAliveV3(targetPid)
		if !isRunning {
			return false, fmt.Errorf("PID [ %d ] exit or got killed, possible err: [ %w ]", targetPid, err)
		}
	}
	return isRunning, nil
}
