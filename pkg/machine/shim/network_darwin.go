//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin && (arm64 || amd64)

package shim

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"bauklotze/pkg/config"
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfigs"

	gvproxy "github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/sirupsen/logrus"
)

const (
	podmanGuestSocks = "/run/podman/podman.sock"
)

func setupMachineSockets(mc *vmconfigs.MachineConfig, _dirs *define.MachineDirs) (string, string, error) {
	host, err := mc.PodmanAPISocketHost()
	if err != nil {
		return "", "", fmt.Errorf("failed to get podman api socket host: %w", err)
	}
	err = host.Delete()
	if err != nil {
		return "", "", fmt.Errorf("failed to delete podman api socket host: %w", err)
	}
	return host.GetPath(), podmanGuestSocks, nil
}

func startHostForwarder(mc *vmconfigs.MachineConfig, provider vmconfigs.VMProvider, dirs *define.MachineDirs, socksInHost string, socksInGuest string) (*exec.Cmd, error) {
	forwardUser := mc.SSH.RemoteUsername

	cfg := config.Default()

	binary, err := cfg.FindHelperBinary(machine.ForwarderBinaryName)
	if err != nil {
		return nil, fmt.Errorf("failed to find helper binary: %w", err)
	}

	cmd := gvproxy.NewGvproxyCommand() // New a GvProxyCommands
	runDir := dirs.RuntimeDir
	logsDIr := dirs.LogsDir

	cmd.PidFile = filepath.Join(runDir.GetPath(), "gvproxy.pid")
	cmd.LogFile = filepath.Join(logsDIr.GetPath(), "gvproxy.log")

	cmd.SSHPort = mc.SSH.Port
	cmd.AddForwardSock(socksInHost)             // podman api in host
	cmd.AddForwardDest(socksInGuest)            // podman api in guest
	cmd.AddForwardUser(forwardUser)             // always be root
	cmd.AddForwardIdentity(mc.SSH.IdentityPath) // ssh keys

	if err := provider.StartNetworking(mc, &cmd); err != nil {
		return nil, fmt.Errorf("failed to start networking: %w", err)
	}

	gvcmd := cmd.Cmd(binary)
	gvcmd.Stdout = os.Stdout
	gvcmd.Stderr = os.Stderr

	if os.Getenv("OVM_DEBUG") == "true" {
		logrus.Infof("Add -debug flag to gvproxy")
		gvcmd.Args = append(gvcmd.Args, "-debug")
	}

	logrus.Infof("Gvproxy command-line: %s", gvcmd.Args)
	if err := gvcmd.Start(); err != nil {
		return nil, fmt.Errorf("unable to execute: %q: %w", cmd.ToCmdline(), err)
	} else {
		machine.GlobalCmds.SetGvpCmd(gvcmd)
	}

	mc.GvProxy.GvProxy.PidFile = cmd.PidFile
	mc.GvProxy.GvProxy.LogFile = cmd.LogFile
	mc.GvProxy.GvProxy.SSHPort = cmd.SSHPort
	mc.GvProxy.GvProxy.MTU = cmd.MTU
	mc.GvProxy.HostSocks = []string{socksInHost}
	mc.GvProxy.RemoteSocks = socksInGuest

	if err := mc.Write(); err != nil {
		return nil, fmt.Errorf("failed to write machine config: %w", err)
	}
	return gvcmd, nil
}
