//go:build (darwin || linux) && (amd64 || arm64)

package gvproxy

import (
	"bauklotze/pkg/machine/machineDefine"
	"errors"
	"fmt"
	psutil "github.com/shirou/gopsutil/v3/process"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"syscall"
	"time"
)

const (
	loops     = 8
	sleepTime = time.Millisecond * 1
)

func waitOnProcess(processID int) error {
	logrus.Infof("Going to stop gvproxy (PID %d)", processID)

	p, err := psutil.NewProcess(int32(processID))
	if err != nil {
		return fmt.Errorf("looking up PID %d: %w", processID, err)
	}

	running, err := p.IsRunning()
	if err != nil {
		return fmt.Errorf("checking if gvproxy is running: %w", err)
	}
	if !running {
		return nil
	}

	if err := p.Kill(); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			logrus.Debugf("Gvproxy already dead, exiting cleanly")
			return nil
		}
		return err
	}
	return backoffForProcess(p)
}

func backoffForProcess(p *psutil.Process) error {
	sleepInterval := sleepTime
	for i := 0; i < loops; i++ {
		running, err := p.IsRunning()
		if err != nil {
			// It is possible that while in our loop, the PID vaporize triggering
			// an input/output error (#21845)
			if errors.Is(err, unix.EIO) {
				return nil
			}
			return fmt.Errorf("checking if process running: %w", err)
		}
		if !running {
			return nil
		}

		time.Sleep(sleepInterval)
		// double the time
		sleepInterval += sleepInterval
	}
	return fmt.Errorf("process %d has not ended", p.Pid)
}

func removeGVProxyPIDFile(f machineDefine.VMFile) error {
	return f.Delete()
}
