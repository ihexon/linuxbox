package ports

import (
	"github.com/sirupsen/logrus"
	"strconv"
	"testing"
)

func TestPorts(t *testing.T) {
	// TODO: This should be in /opt/ovmd, but for now just `echo OkImFine`
	newPort, err := AllocateMachinePort()
	if err != nil {
		t.Errorf(err.Error())
	}
	t.Logf(strconv.Itoa(newPort))

	success := false
	defer func() {
		if !success {
			if err := ReleaseMachinePort(newPort); err != nil {
				logrus.Warnf("could not release port allocation as part of failure rollback (%d): %s", newPort, err.Error())
			}
		}
	}()

}
