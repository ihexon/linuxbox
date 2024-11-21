package system

import (
	"os"
	"os/exec"
	"strconv"
	"testing"
)

func TestIsProcessAlive(t *testing.T) {
	// get pid list current system using ps -e
	// ps -e | grep usr| grep -v  grep | cut -d ' ' -f1| xargs| sed 's/ /,/g'
	isRunning, err := IsProcesSAlive([]int32{11250, 11446, 12410, 12419, 12450, 12451, 12453, 12454, 12458, 12462, 12605, 12608, 12609, 12754, 13872, 16753, 17188, 21257, 22521, 26055, 26062, 33612, 33613, 37284, 38093, 43486, 43487, 43951, 45245, 45246, 45552, 46159, 48412, 48770, 52906, 54849, 54851, 58239, 68594, 86290, 97027})
	if isRunning {
		t.Logf("all pids alive")
	} else {
		t.Errorf("some pids not alive, err: %v", err)
	}
}

func TestGetPPID(t *testing.T) {
	pid := os.Getpid()
	t.Logf("PID is: %d", pid)
	ppid, err := GetPPID(int32(pid))
	if err != nil {
		t.Errorf("GetPPID failed, err: %v", err)
	}
	t.Logf("PPID is: %d", ppid)

	// Using ps -p <pid> -o ppid= find ppid of pid
	t.Logf("Using ps get ppid : ps -p %d -o ppid=", ppid)
	s, _ := exec.Command("ps", "-p", strconv.Itoa(int(pid)), "-o", "ppid=").Output()
	t.Logf("PPID is: %s", s)

}
