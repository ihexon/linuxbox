package shim

import (
	"bauklotze/pkg/config"
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/connection"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/machineDefine"
	"bauklotze/pkg/machine/ports"
	"bauklotze/pkg/machine/vmconfigs"
	"fmt"
	gvproxy "github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

const (
	defaultGuestSock = "/run/user/%d/podman/podman.sock"
)

func reassignSSHPort(mc *vmconfigs.MachineConfig, provider vmconfigs.VMProvider) error {
	newPort, err := ports.AllocateMachinePort()
	if err != nil {
		return err
	}

	success := false
	defer func() {
		if !success {
			if err := ports.ReleaseMachinePort(newPort); err != nil {
				logrus.Warnf("could not release port allocation as part of failure rollback (%d): %s", newPort, err.Error())
			}
		}
	}()

	// Write a transient invalid port, to force a retry on failure
	oldPort := mc.SSH.Port
	mc.SSH.Port = 0
	if err := mc.Write(); err != nil {
		return err
	}

	if err := ports.ReleaseMachinePort(oldPort); err != nil {
		logrus.Warnf("could not release current ssh port allocation (%d): %s", oldPort, err.Error())
	}

	// Update the backend's settings if relevant (e.g. WSL)
	if err := provider.UpdateSSHPort(mc, newPort); err != nil {
		return err
	}

	mc.SSH.Port = newPort
	if err := connection.UpdateConnectionPairPort(mc.Name, newPort, mc.HostUser.UID, mc.SSH.RemoteUsername, mc.SSH.IdentityPath); err != nil {
		return fmt.Errorf("could not update remote connection configuration: %w", err)
	}

	// Write updated port back
	if err := mc.Write(); err != nil {
		return err
	}

	// inform defer routine not to release the port
	success = true

	return nil
}

func startHostForwarder(mc *vmconfigs.MachineConfig, provider vmconfigs.VMProvider, dirs *machineDefine.MachineDirs, hostSocks []string) error {
	forwardUser := mc.SSH.RemoteUsername

	// TODO should this go up the stack higher or
	// the guestSock is "inside" the guest machine
	guestSock := fmt.Sprintf(defaultGuestSock, mc.HostUser.UID)
	if mc.HostUser.Rootful {
		guestSock = "/run/podman/podman.sock"
		forwardUser = "root"
	}

	cfg, err := config.Default()
	if err != nil {
		return err
	}

	binary, err := cfg.FindHelperBinary(machine.ForwarderBinaryName, false)
	if err != nil {
		return err
	}
	cmd := gvproxy.NewGvproxyCommand()
	runDir := dirs.RuntimeDir
	cmd.PidFile = filepath.Join(runDir.GetPath(), "gvproxy.pid")
	cmd.LogFile = filepath.Join(runDir.GetPath(), "gvproxy.log")
	cmd.SSHPort = mc.SSH.Port

	for _, hostSock := range hostSocks {
		cmd.AddForwardSock(hostSock)
		cmd.AddForwardDest(guestSock)
		cmd.AddForwardUser(forwardUser)
		cmd.AddForwardIdentity(mc.SSH.IdentityPath)
	}

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		cmd.Debug = true
		logrus.Debug(cmd)
	}

	if err := provider.StartNetworking(mc, &cmd); err != nil {
		return err
	}
	c := cmd.Cmd(binary)

	logrus.Debugf("gvproxy command-line: %s %s", binary, strings.Join(cmd.ToCmdline(), " "))
	if err := c.Start(); err != nil {
		return fmt.Errorf("unable to execute: %q: %w", cmd.ToCmdline(), err)
	}
	return nil
}

func startNetworking(mc *vmconfigs.MachineConfig, provider vmconfigs.VMProvider) (string, machine.APIForwardingState, error) {
	// Check if SSH port is in use, and reassign if necessary
	if !ports.IsLocalPortAvailable(mc.SSH.Port) {
		logrus.Warnf("detected port conflict on machine ssh port [%d], reassigning", mc.SSH.Port)
		if err := reassignSSHPort(mc, provider); err != nil {
			return "", 0, err
		}
	}

	// Provider has its own networking code path (e.g. WSL)
	if provider.UseProviderNetworkSetup() {
		return "", machine.NoForwarding, provider.StartNetworking(mc, nil)
	}

	dirs, err := env.GetMachineDirs(provider.VMType())
	if err != nil {
		return "", machine.NoForwarding, err
	}

	hostSocks, forwardSock, forwardingState, err := setupMachineSockets(mc, dirs)
	if err != nil {
		return "", machine.NoForwarding, err
	}

	if err := startHostForwarder(mc, provider, dirs, hostSocks); err != nil {
		return "", machine.NoForwarding, err
	}

	return forwardSock, forwardingState, nil
}
