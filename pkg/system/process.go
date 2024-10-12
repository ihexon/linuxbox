package system

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sys/unix"
	"os"
	"syscall"
)

// IsProcessAlive returns true if process with a given pid is running.
func IsProcessAlive(pid int) bool {
	err := unix.Kill(pid, syscall.Signal(0))
	if err == nil || err == unix.EPERM {
		return true
	}
	return false
}

// KillProcess force-stops a process.
func KillProcess(pid int) {
	_ = unix.Kill(pid, unix.SIGKILL)
}
func CheckProcessRunning(processName string, pid int) error {
	var status syscall.WaitStatus
	// wait(): on success, returns the process ID of the terminated
	// child; on failure, -1 is returned.
	pid, err := syscall.Wait4(pid, &status, syscall.WNOHANG, nil)
	if err != nil {
		return fmt.Errorf("failed to read %s process status: %w", processName, err)
	}
	if pid > 0 {
		// child exited
		return fmt.Errorf("%s exited unexpectedly with exit code %d", processName, status.ExitStatus())
	}
	return nil
}

func GetMyPPID() (int32, error) {
	pid := os.Getpid()
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
