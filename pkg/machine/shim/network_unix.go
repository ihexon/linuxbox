//go:build dragonfly || freebsd || linux || netbsd || openbsd || darwin

package shim

import (
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/machineDefine"
	"bauklotze/pkg/machine/vmconfigs"
)

func setupMachineSockets(mc *vmconfigs.MachineConfig, dirs *machineDefine.MachineDirs) ([]string, string, machine.APIForwardingState, error) {
	// podman-api.sock in guest --> podman-api.sock in host
	hostSocket, err := mc.APISocket()
	if err != nil {
		return nil, "", machine.NoForwarding, err
	}

	forwardSock, state, err := setupForwardingLinks(hostSocket, dirs.DataDir)
	if err != nil {
		return nil, "", machine.NoForwarding, err
	}
	return []string{hostSocket.GetPath()}, forwardSock, state, err
}

func setupForwardingLinks(hostSocket, dataDir *machineDefine.VMFile) (string, machine.APIForwardingState, error) {
	return hostSocket.GetPath(), machine.NotInstalled, nil
}
