//go:build darwin && (arm64 || amd64)

package system

import "github.com/shirou/gopsutil/v3/process"

// IsProcessAliveV3 returns true if process with a given pid is running.

func IsProcessAliveV3(pid int32) (bool, error) {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return false, err
	}
	s, err := proc.Status()
	if err != nil {
		return false, err
	}

	for _, v := range s {
		if v != process.Zombie {
			return true, nil
		}
	}
	return false, err
}
