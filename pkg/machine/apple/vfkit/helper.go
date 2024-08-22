package vfkit

import (
	"bauklotze/pkg/config"
	"bauklotze/pkg/machine/define"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"time"
)

// Helper describes the use of vfkit: cmdline and endpoint
type Helper struct {
	LogLevel       logrus.Level
	Endpoint       string
	BinaryPath     *define.VMFile
	VirtualMachine *config.VirtualMachine
	Rosetta        bool
}

// TODO: vfkit.Stop
func (vf *Helper) Stop(force, wait bool) error {
	state := rest.Stop
	if force {
		state = rest.HardStop
	}
	if err := vf.stateChange(state); err != nil {
		return err
	}
	if !wait {
		return nil
	}
	waitDuration := time.Millisecond * 500
	// Wait up to 90s then hard force off
	for i := 0; i < 180; i++ {
		_, err := vf.getRawState()
		if err != nil || errors.Is(err, unix.ECONNREFUSED) {
			return nil
		}
		time.Sleep(waitDuration)
	}
	logrus.Warn("Failed to gracefully stop machine, performing hard stop")
	// we waited long enough do a hard stop
	return vf.stateChange(rest.HardStop)
}
