//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package shim

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"time"
	"unsafe"

	gvproxy "github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/containers/storage/pkg/fileutils"
	"github.com/sirupsen/logrus"

	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/events"
	"bauklotze/pkg/machine/vmconfig"
	"bauklotze/pkg/system"
)

// podmanSockets returns the path to the podman api socket on host/guest
func setupPodmanSocketsPath(mc *vmconfig.MachineConfig) (string, string, error) {
	podmanOnHost := mc.PodmanAPISocketHost()
	if err := podmanOnHost.Delete(true); err != nil {
		return "", "", fmt.Errorf("failed to delete podman api socket host: %w", err)
	}
	return podmanOnHost.GetPath(), define.PodmanGuestSocks, nil
}

func tryKillGvProxyBeforRun(mc *vmconfig.MachineConfig) {
	gvpBin := mc.Dirs.NetworkProvider.Bin
	proc, _ := system.FindProcessByPath(gvpBin.GetPath())
	if proc != nil {
		logrus.Warnf("Find running %s process, this should never happen, try to kill", gvpBin.GetPath())
		_ = proc.Kill()
	}
}

func startForwarder(ctx context.Context, mc *vmconfig.MachineConfig, socksOnHost string, socksOnGuest string) error {
	tryKillGvProxyBeforRun(mc)
	gvpBin := mc.Dirs.NetworkProvider.Bin
	logrus.Infof("Gvproxy binary: %s", mc.Dirs.NetworkProvider.Bin.GetPath())
	if !gvpBin.Exist() {
		return fmt.Errorf("%s does not exist", gvpBin.GetPath())
	}

	socksDir := mc.Dirs.SocksDir
	gvproxyCommand := gvproxy.NewGvproxyCommand() // New a GvProxyCommands

	gvpPidFile, _ := socksDir.AppendToNewVMFile(fmt.Sprintf("%s-%s", mc.VMName, define.GvProxyPidName))
	if err := gvpPidFile.Delete(true); err != nil {
		return fmt.Errorf("failed to delete pid file: %w", err)
	}
	gvproxyCommand.PidFile = gvpPidFile.GetPath()
	gvproxyCommand.SSHPort = mc.SSH.Port
	gvproxyCommand.AddForwardSock(socksOnHost)             // podman api in host
	gvproxyCommand.AddForwardDest(socksOnGuest)            // podman api in guest
	gvproxyCommand.AddForwardUser(mc.SSH.RemoteUsername)   // always be root
	gvproxyCommand.AddForwardIdentity(mc.SSH.IdentityPath) // ssh keys

	// This allows a provider to perform additional setup cause vfkit/krunkit are different hypervisor
	// and have different networking configure
	if err := mc.VMProvider.SetupProviderNetworking(mc, &gvproxyCommand); err != nil {
		return fmt.Errorf("failed to setup provider networking: %w", err)
	}

	if os.Getenv("OVM_GVPROXY_DEBUG") == "true" {
		logrus.Warn("gvproxy running in debug mode")
		gvproxyCommand.Debug = true
	}

	v := reflect.ValueOf(&gvproxyCommand).Elem().FieldByName("forwardInfo")
	aArray := reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Interface().(map[string][]string)
	mc.GvProxy.ForwardInfo = aArray

	v = reflect.ValueOf(&gvproxyCommand).Elem().FieldByName("sockets")
	bArray := reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Interface().(map[string]string)
	mc.GvProxy.Sockets = bArray

	gvpExecCmd := exec.CommandContext(ctx, gvpBin.GetPath(), gvproxyCommand.ToCmdline()...)
	gvpExecCmd.Stdout = os.Stdout
	gvpExecCmd.Stderr = os.Stderr

	logrus.Infof("GVPROXY FULL CMDLINE: %q", gvpExecCmd.Args)
	events.NotifyRun(events.StartGvProxy, "staring...")
	if err := gvpExecCmd.Start(); err != nil {
		return fmt.Errorf("unable to execute: %q: %w", gvpExecCmd.Args, err)
	}

	events.NotifyRun(events.StartGvProxy, "finished")

	mc.GvProxy.HostSocks = []string{socksOnHost}
	mc.GvProxy.PidFile = gvproxyCommand.PidFile
	mc.GvProxy.SSHPort = gvproxyCommand.SSHPort
	mc.GvProxy.MTU = gvproxyCommand.MTU
	mc.GvpCmd = gvpExecCmd

	socks, _ := mc.GVProxyNetworkBackendSocks()

	// WaitForSocket when gvproxy started, we also check that the gvproxy socket is created
	// there is a little chance that the socket is not created, causing krunkit to freeze
	if err := waitForSocket(ctx, socks.GetPath()); err != nil {
		return fmt.Errorf("failed to wait for gvproxy network backend socks: %w", err)
	}

	if err := mc.Write(); err != nil {
		return fmt.Errorf("failed to write machine config: %w", err)
	}
	return nil
}

func waitForSocket(ctx context.Context, socketPath string) error {
	var backoff = 100 * time.Millisecond
	logrus.Infof("Test that %s socket is created", socketPath)
	// we wait for the socket to be created, when gvproxy first run on macOS
	// the Gatekeeper/Notarization will slow done the gvproxy code executed
	for range 100 {
		select {
		case <-ctx.Done():
			return fmt.Errorf("cancel waitForSocket,ctx cancelled: %w", context.Cause(ctx))
		default:
			if err := fileutils.Exists(socketPath); err != nil {
				logrus.Warnf("Gvproxy network backend socket not ready, try test %s again....", socketPath)
				time.Sleep(backoff)
			}
			return nil
		}
	}
	return fmt.Errorf("gvproxy network backend socket file not created: %s", socketPath)
}
