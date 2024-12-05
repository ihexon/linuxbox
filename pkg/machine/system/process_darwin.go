//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin && (arm64 || amd64)

package system

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/process"
)

// IsProcessAliveV3 returns true if process with a given pid is running.
func IsProcessAliveV3(pid int32) (bool, error) {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return false, fmt.Errorf("failed to find process: %w", err)
	}
	s, err := proc.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get process status: %w", err)
	}

	for _, v := range s {
		switch v {
		case process.Zombie:
		case process.Stop:
		case process.UnknownState:
			return false, nil
		default:
			return true, nil
		}
	}
	return false, nil
}

func KillProcess(pid int) error {
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}
	_ = proc.Kill()

	return nil
}
