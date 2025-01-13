//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package shim

import (
	"errors"
	"os/exec"
	"time"

	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/vmconfigs"

	"github.com/sirupsen/logrus"
)

var (
	defaultBackoff     = 100 * time.Millisecond
	maxTried           = 200
	ErrNotRunning      = errors.New("machine not in running state")
	ErrSSHNotListening = errors.New("machine is not listening on ssh port")
)

// conductVMReadinessCheck checks to make sure SSH is up and running
func conductVMReadinessCheck(mc *vmconfigs.MachineConfig) bool {
	for i := range maxTried {
		if i > 0 {
			time.Sleep(defaultBackoff)
		}

		if err := machine.CommonSSHSilent(mc.SSH.RemoteUsername, mc.SSH.IdentityPath, mc.Name, mc.SSH.Port, []string{"echo Hello"}); err != nil {
			logrus.Warnf("SSH readiness check for machine failed: %v", err)
			continue
		}
		return true
	}
	return false
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
