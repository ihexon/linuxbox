//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package shim

import (
	"context"
	"fmt"
	"time"

	sshService "bauklotze/pkg/machine/ssh/service"

	"github.com/sirupsen/logrus"

	"bauklotze/pkg/machine/io"
	"bauklotze/pkg/machine/vmconfig"
)

var (
	defaultBackoff = 100 * time.Millisecond
	maxTried       = 100
)

// ConductVMReadinessCheck checks to make sure SSH is up and running
func ConductVMReadinessCheck(ctx context.Context, mc *vmconfig.MachineConfig) bool {
	for i := range maxTried {
		if ctx.Err() != nil {
			return false
		}

		if i > 0 {
			time.Sleep(defaultBackoff)
		}

		if err := sshService.GetKernelInfo(ctx, mc); err != nil {
			logrus.Warnf("SSH readiness check for machine failed: %v, try again", err)
			continue
		}
		return true
	}
	return false
}

// startNetworking return podman socks in host, podman socks in guest, error
func startNetworking(ctx context.Context, mc *vmconfig.MachineConfig) (*io.FileWrapper, *io.FileWrapper, error) {
	// socksInHost($workspace/tmp/[machine]-podman-api.socks) <--> socksInGuest(podman server)
	socksInHost, socksInGuest, err := setupPodmanSocketsPath(mc)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup podman sockets path: %w", err)
	}

	err = startForwarder(ctx, mc, socksInHost, socksInGuest)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start forwarder: %w", err)
	}

	return &io.FileWrapper{Path: socksInHost}, &io.FileWrapper{Path: socksInGuest}, err
}
