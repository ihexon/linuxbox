//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package krunkit

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"bauklotze/pkg/machine/define"
	io2 "bauklotze/pkg/machine/io"
	"bauklotze/pkg/machine/vmconfig"
	"bauklotze/pkg/registry"
	"bauklotze/pkg/system"

	gvproxyTypes "github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/containers/storage/pkg/fileutils"
	"github.com/sirupsen/logrus"
)

func startGvproxy(ctx context.Context, mc *vmconfig.MachineConfig) error {
	if err := system.KillExpectProcNameFromPPIDFile(mc.PIDFiles.GvproxyPidFile, define.GvProxyBinaryName); err != nil {
		logrus.Warnf("kill old gvproxy from pid process failed: %v", err)
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("unable to get executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("unable to eval symlinks: %w", err)
	}

	gvpBin := filepath.Join(filepath.Dir(filepath.Dir(execPath)), define.Libexec, define.GvProxyBinaryName)
	gvpCmd := gvproxyTypes.NewGvproxyCommand()
	gvpCmd.SSHPort = mc.SSH.Port
	// gvproxy listen a local socks file as Podman API socks (PodmanSocks.InHost)
	// and forward to the guest's Podman API socks(PodmanSocks.InGuest).
	gvpCmd.AddForwardSock(mc.PodmanSocks.InHost)
	gvpCmd.AddForwardDest(mc.PodmanSocks.InGuest)
	gvpCmd.AddForwardUser(mc.SSH.RemoteUsername)
	gvpCmd.AddForwardIdentity(mc.SSH.PrivateKey)
	gvpCmd.PidFile = mc.PIDFiles.GvproxyPidFile

	// gvproxy endpoint, which provide network backend for vfkit/krunkit
	gvpEndPoint := io2.NewFile(mc.GetNetworkStackEndpoint())

	if err := gvpEndPoint.Delete(true); err != nil {
		return fmt.Errorf("unable to remove gvproxy endpoint file: %w", err)
	}
	gvpCmd.AddVfkitSocket(fmt.Sprintf("unixgram://%s", gvpEndPoint.GetPath()))

	if os.Getenv("OVM_GVPROXY_DEBUG") == "true" {
		logrus.Warn("gvproxy running in debug mode")
		gvpCmd.Debug = true
	}

	cmd := exec.CommandContext(ctx, gvpBin, gvpCmd.ToCmdline()...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	logrus.Infof("GVPROXY FULL CMDLINE: %q", cmd.Args)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to execute: %q: %w", cmd.Args, err)
	}

	defer registry.RegistryCmd(cmd)

	return waitForSocket(ctx, gvpEndPoint.GetPath())
}

// we wait for the socket to be created, when gvproxy first run on macOS
// the Gatekeeper/Notarization will slow done the gvproxy code executed
func waitForSocket(ctx context.Context, socketPath string) error {
	var backoff = 100 * time.Millisecond
	for range 100 {
		select {
		case <-ctx.Done():
			return fmt.Errorf("cancel waitForSocket,ctx cancelled: %w", context.Cause(ctx))
		default:
			if err := fileutils.Exists(socketPath); err != nil {
				logrus.Warnf("Gvproxy network backend socket not ready, try test %q again....", socketPath)
				time.Sleep(backoff)
				continue
			}
			return nil
		}
	}
	return fmt.Errorf("gvproxy network backend socket file not created in %q", socketPath)
}
