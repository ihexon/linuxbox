package shim

import (
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/connection"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/ports"
	"bauklotze/pkg/machine/vmconfigs"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

const (
	defaultGuestSock = "/run/user/%d/podman/podman.sock"
)

var (
	ErrNotRunning      = errors.New("machine not in running state")
	ErrSSHNotListening = errors.New("machine is not listening on ssh port")
)

func isListening(port int) bool {
	// Check if we can dial it
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", port), 10*time.Millisecond)
	if err != nil {
		return false
	}
	if err := conn.Close(); err != nil {
		logrus.Error(err)
	}
	return true
}

// conductVMReadinessCheck checks to make sure the machine is in the proper state
// and that SSH is up and running
func conductVMReadinessCheck(mc *vmconfigs.MachineConfig, maxBackoffs int, backoff time.Duration, stateF func() (define.Status, error)) (connected bool, sshError error, err error) {
	for i := 0; i < maxBackoffs; i++ {
		if i > 0 {
			time.Sleep(backoff)
			backoff *= 2
		}
		state, err := stateF()
		if err != nil {
			return false, nil, err
		}
		if state != define.Running {
			sshError = ErrNotRunning
			continue
		}
		if !isListening(mc.SSH.Port) {
			sshError = ErrSSHNotListening
			continue
		}
		if sshError = machine.CommonSSHSilent(mc.SSH.RemoteUsername, mc.SSH.IdentityPath, mc.Name, mc.SSH.Port, []string{"true"}); sshError != nil {
			logrus.Debugf("SSH readiness check for machine failed: %v", sshError)
			continue
		}
		connected = true
		sshError = nil
		break
	}
	return
}

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

	logrus.Infof("Update ssh port for %s, new ssh port: %s", mc.Name, newPort)
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
			return "", machine.NoForwarding, err
		}
	}

	dirs, err := env.GetMachineDirs(provider.VMType())
	if err != nil {
		return "", machine.NoForwarding, err
	}

	hostSocks, forwardSock, _, err := setupMachineSockets(mc, dirs)
	if err != nil {
		return "", machine.NoForwarding, err
	}

	if err := startHostForwarder(mc, provider, dirs, hostSocks); err != nil {
		return "", machine.NoForwarding, err
	}

	return forwardSock, machine.InForwarding, nil
}
