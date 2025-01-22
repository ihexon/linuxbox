package system

import "testing"

func TestIsProcessAliveV4(t *testing.T) {
	_, err := IsProcessAliveV4(1)
	if err != nil {
		t.Errorf("IsAlive(1) error: %v", err)
	}
	_, err = IsProcessAliveV4(65533)
	if err != nil {
		t.Errorf("IsAlive(1) error: %v", err)
	}
}

func TestFindPidByPath(t *testing.T) {
	proc, err := FindProcessByPath("/sbin/launchd")
	if err != nil {
		t.Errorf("findProcessByPath(/sbin/launchd) error: %v", err)
	}
	t.Logf("found process with pid %d", proc.Pid)
}
