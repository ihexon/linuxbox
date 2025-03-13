//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"bauklotze/pkg/machine/vmconfig"
)

func GetKernelInfo(ctx context.Context, mc *vmconfig.MachineConfig) error {
	return runCtx(ctx, mc, "uname", []string{
		"-a",
	})
}

func DoTimeSync(ctx context.Context, mc *vmconfig.MachineConfig) error {
	return runCtx(ctx, mc, "date", []string{
		"-s",
		fmt.Sprintf("@%d", time.Now().Unix()),
	})
}

func DoSync(mc *vmconfig.MachineConfig) error {
	return run(mc, "bash", []string{
		"-c",
		"sync",
	})
}

func GracefulShutdownVK(mc *vmconfig.MachineConfig) error {
	logrus.Infoln("stop all containers")
	if err := run(mc, "podman", []string{
		"stop",
		"-a",
		"-t",
		"3",
	}); err != nil {
		logrus.Warnf("podman stop failed: %v", err)
	}

	logrus.Infoln("sync disk")
	if err := run(mc, "sync", []string{}); err != nil {
		logrus.Warnf("sync disk failed: %v", err)
	}

	logrus.Infoln("stop vm now")
	// poweroff (provided by busybox init system) cause vCPU 0 received shutdown signal,and active power off
	// so the krunkit will exit clearly after vm shutdown
	// halt (provided by busybox init system) shutdown the kernel, but uninterrupted power supply, so the krunkit
	// will continue to run after vm shutdown
	if err := run(mc, "poweroff", []string{}); err != nil {
		return fmt.Errorf("stop vm failed: %w", err)
	}

	return nil
}
