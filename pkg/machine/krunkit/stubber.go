//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package krunkit

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/events"
	"bauklotze/pkg/machine/network"
	"bauklotze/pkg/machine/ssh/service"
	"bauklotze/pkg/system"

	"bauklotze/pkg/machine/vmconfig"
	"bauklotze/pkg/pty"
	"bauklotze/pkg/registry"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Stubber struct {
	VMState *vmconfig.VMState
}

func NewProvider() *Stubber {
	return &Stubber{
		VMState: &vmconfig.VMState{
			SSHReady:    false,
			PodmanReady: false,
		},
	}
}

func (l *Stubber) VMType() vmconfig.VMType {
	return vmconfig.LibKrun
}

func (l *Stubber) StartNetworkProvider(ctx context.Context, mc *vmconfig.MachineConfig) error {
	if err := network.StartGvproxy(ctx, mc); err != nil {
		return fmt.Errorf("failed to start network stack: %w", err)
	}

	logrus.Infof("network stack started")
	return nil
}

func (l *Stubber) StartVMProvider(ctx context.Context, mc *vmconfig.MachineConfig) error {
	if err := startKrunKit(ctx, mc); err != nil {
		return fmt.Errorf("failed to start virtual machine: %w", err)
	}

	if machine.WaitSSHStarted(ctx, mc) {
		logrus.Infof("vm ssh service started")
	}
	l.VMState.SSHReady = true

	if err := machine.WaitPodmanReady(ctx, mc.PodmanSocks.InHost); err != nil {
		logrus.Infof("vm podman service started")
	}
	l.VMState.PodmanReady = true
	events.NotifyRun(events.Ready)

	return nil
}

func (l *Stubber) StartSSHAuthService(ctx context.Context, mc *vmconfig.MachineConfig) error {
	sshAuthService := service.NewSSHAuthService(
		mc.SSHAuthSocks.LocalSocks,
		mc.SSHAuthSocks.RemoteSocks,
		mc.SSH.RemoteUsername,
		mc.SSH.PrivateKey,
		mc.SSH.Port,
	)

	g, ctx2 := errgroup.WithContext(ctx)
	g.Go(func() error {
		return sshAuthService.StartSSHAuthServiceAndForwardV2(ctx2)
	})

	g.Go(func() error {
		return sshAuthService.StartUnixSocketForward(ctx2)
	})

	return g.Wait() //nolint:wrapcheck
}

func (l *Stubber) StartTimeSyncService(ctx context.Context, mc *vmconfig.MachineConfig) error {
	return machine.SyncTimeOnWake(ctx, mc) //nolint:wrapcheck
}

func startKrunKit(ctx context.Context, mc *vmconfig.MachineConfig) error {
	if err := system.KillExpectProcNameFromPPIDFile(mc.PIDFiles.KrunKitPidFile, define.KrunkitBinaryName); err != nil {
		logrus.Warnf("kill krunkit from pid process failed: %v", err)
	}

	vmc, err := machine.CreateDynamicConfigure(mc)
	if err != nil {
		return fmt.Errorf("failed to create dynamic machine configure: %w", err)
	}

	cmd, err := vmc.Cmd(mc.KrunKitBin)
	if err != nil {
		return fmt.Errorf("failed to create krunkit command: %w", err)
	}

	cmd.Args = append(cmd.Args, "--krun-log-level", "3")

	cmd = exec.CommandContext(ctx, mc.KrunKitBin, cmd.Args[1:]...)

	logrus.Infof("full cmdline: %q", cmd.Args)

	events.NotifyRun(events.StartKrunKit)
	ptmx, err := pty.RunInPty(cmd)
	if err != nil {
		return fmt.Errorf("failed to run krunkit in pty: %w", err)
	}

	go func() {
		_, _ = io.Copy(os.Stdout, ptmx)
	}()

	if err := os.WriteFile(mc.PIDFiles.KrunKitPidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644); err != nil {
		return fmt.Errorf("unable to write krunkit pid file: %w", err)
	}

	registry.RegistryCmd(cmd)
	return nil
}

func (l *Stubber) InitializeVM(opts vmconfig.VMOpts) (*vmconfig.MachineConfig, error) {
	return machine.InitializeVM(opts) //nolint:wrapcheck
}

func (l *Stubber) GetVMState() *vmconfig.VMState {
	return l.VMState
}
