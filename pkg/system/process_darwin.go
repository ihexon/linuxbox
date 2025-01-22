//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package system

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/process"
)

// IsProcessAliveV4 if return err != nil, process not found
func IsProcessAliveV4(pid int) (bool, error) {
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return false, fmt.Errorf("failed to find process: %w", err)
	}
	running, err := proc.IsRunning()
	if err != nil {
		return false, fmt.Errorf("failed to check if process is running: %w", err)
	}
	if !running {
		return false, fmt.Errorf("process %d not found", pid)
	}
	return true, nil
}
