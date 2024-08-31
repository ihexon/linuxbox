package shim

import (
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/connection"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/ports"
	"bauklotze/pkg/machine/vmconfigs"
	"fmt"
	"github.com/sirupsen/logrus"
)

const (
	defaultGuestSock = "/run/user/%d/ovm/ovm_guest.sock"
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
		return "", 0, provider.StartNetworking(mc, nil)
	}

	dirs, err := env.GetMachineDirs(provider.VMType())
	if err != nil {
		return "", 0, err
	}

	hostSocks, forwardSock, forwardingState, err := setupMachineSockets(mc, dirs)
	if err != nil {
		return "", 0, err
	}

	if err := startHostForwarder(mc, provider, dirs, hostSocks); err != nil {
		return "", 0, err
	}

	return forwardSock, forwardingState, nil
}
