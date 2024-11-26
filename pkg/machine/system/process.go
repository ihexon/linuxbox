package system

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/process"
)

func GetPPID(pid int32) (int32, error) {
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return -1, err
	}
	ppid, err := proc.Ppid()
	if err != nil {
		return -1, err
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
			return false, fmt.Errorf("PID [ %d ] exit or got killed, possible err: [ %v ]", targetPid, err)
		}
	}
	return isRunning, nil
}
