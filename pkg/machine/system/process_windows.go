//go:build windows && (arm64 || amd64)

package system

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/process"
)

func IsProcessAliveV3(pid int32) (bool, error) {
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return false, err
	}
	isRunning, err := proc.IsRunning()
	if err != nil {
		return false, err
	}

	return isRunning, err
}
func KillProcess(pid int) error {
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return fmt.Errorf("failed to find process: %v", err)
	}

	err = proc.Terminate()
	if err != nil {
		return fmt.Errorf("failed to kill process: %v", err)
	}

	return nil
}
