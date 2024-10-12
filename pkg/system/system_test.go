package system

import "testing"

func TestIsProcessAliveV2(t *testing.T) {
	pid := 212
	isRunning, err := IsProcessAliveV2(pid)
	if err != nil {
		t.Errorf(err.Error())
	}
	if isRunning {
		t.Logf("Process %d is  running", pid)
	} else {
		t.Logf("Process %d is not running", pid)
	}
}
