package system

import "testing"

func TestFindPidByCommandLine(t *testing.T) {
	cmdline, err := FindPIDByCmdline("ovm/vm-res")
	if err != nil {
		t.Logf("error: %v", err)
	}
	for _, pid := range cmdline {
		t.Logf("pid: %d", pid)
	}
}
