//go:build windows && (arm64 || amd64)

package system

func IsProcessAliveV3(pid int) (bool, error) {
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
