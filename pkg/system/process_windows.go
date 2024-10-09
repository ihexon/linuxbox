//go:build windows && (arm64 || amd64)

package system

import (
	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sys/windows"
)

// TODO process_windows
func IsProcessAlive(pid int) bool {
	panic("not implement...")
	return true
}

// TODO process_windows
func CheckProcessRunning(processName string, pid int) error {
	panic("not implement...")
	return nil
}

func GetMyPPID() (int32, error) {
	pid := windows.Getpid()
	ppid, err := GetPidOfPPID(int32(pid))
	if err != nil {
		return ppid, err
	}
	return ppid, nil
}

func GetPidOfPPID(pid int32) (int32, error) {
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
