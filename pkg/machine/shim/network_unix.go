//go:build dragonfly || freebsd || linux || netbsd || openbsd || darwin

package shim

import (
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfigs"
)

func setupMachineSockets(mc *vmconfigs.MachineConfig, dirs *define.MachineDirs) ([]string, string, machine.APIForwardingState, error) {
	hostSocket, err := mc.APISocket()
	if err != nil {
		return nil, "", 0, err
	}

	forwardSock, state, err := setupForwardingLinks(hostSocket, dirs.DataDir)
	if err != nil {
		return nil, "", 0, err
	}
	return []string{hostSocket.GetPath()}, forwardSock, state, nil
}

func setupForwardingLinks(hostSocket, dataDir *define.VMFile) (string, machine.APIForwardingState, error) {
	_ = hostSocket.Delete()
	return hostSocket.GetPath(), machine.NotInstalled, nil
}
