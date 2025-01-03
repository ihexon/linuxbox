//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package shim

import (
	"errors"
	"fmt"
	"os/exec"
	"time"

	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/port"

	"github.com/sirupsen/logrus"
)

var (
	defaultBackoff     = 100 * time.Millisecond
	maxTried           = 200
	ErrNotRunning      = errors.New("machine not in running state")
	ErrSSHNotListening = errors.New("machine is not listening on ssh port")
)

// conductVMReadinessCheck checks to make sure the machine is in the proper state
// and that SSH is up and running
func conductVMReadinessCheck(mc *vmconfigs.MachineConfig, stateF func() (define.Status, error)) (connected bool, sshError error, err error) {
	for i := range maxTried {
		if i > 0 {
			time.Sleep(defaultBackoff)
		}
		state, err := stateF()
		if err != nil {
			return false, nil, fmt.Errorf("failed to get machine state: %w", err)
		}

		if state != define.Running {
			sshError = ErrNotRunning
			continue
		}

		if !port.IsListening(mc.SSH.Port) {
			sshError = ErrSSHNotListening
			continue
		}

		if sshError = machine.CommonSSHSilent(mc.SSH.RemoteUsername, mc.SSH.IdentityPath, mc.Name, mc.SSH.Port, []string{"echo Hello"}); sshError != nil {
			logrus.Warnf("SSH readiness check for machine failed: %v", sshError)
			continue
		}
		connected = true
		sshError = nil
		break
	}
	return
}

func startNetworking(mc *vmconfigs.MachineConfig, provider vmconfigs.VMProvider) (string, machine.APIForwardingState, *exec.Cmd, error) {
	socksInHost, socksInGuest, err := setupMachineSockets(mc, mc.Dirs)
	if err != nil {
		return "", machine.NoForwarding, nil, err
	}

	// forward the IO in socksInHost to socksInGuest
	gvcmd, err := startHostForwarder(mc, provider, mc.Dirs, socksInHost, socksInGuest)
	if err != nil {
		return "", machine.NoForwarding, nil, err
	}

	return socksInHost, machine.InForwarding, gvcmd, nil
}
