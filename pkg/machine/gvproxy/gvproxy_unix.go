//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build (darwin || linux) && (amd64 || arm64)

package gvproxy

import (
	"errors"
	"fmt"
	"syscall"
	"time"

	"bauklotze/pkg/machine/define"

	psutil "github.com/shirou/gopsutil/v3/process"
	"github.com/sirupsen/logrus"
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
	return nil
}

func removeGVProxyPIDFile(f define.VMFile) error {
	return f.Delete()
}
