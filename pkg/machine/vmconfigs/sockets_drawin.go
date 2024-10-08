package vmconfigs

import (
	"bauklotze/pkg/machine/define"
	"fmt"
)

func readySocket(name string, machineRuntimeDir *define.VMFile) (*define.VMFile, error) {
	socketName := fmt.Sprintf("%s-ready.sock", name)
	return machineRuntimeDir.AppendToNewVMFile(socketName, nil)
}

func gvProxySocket(name string, machineRuntimeDir *define.VMFile) (*define.VMFile, error) {
	socketName := fmt.Sprintf("%s-gvproxy.sock", name)
	return machineRuntimeDir.AppendToNewVMFile(socketName, nil)
}

func apiSocket(name string, socketDir *define.VMFile) (*define.VMFile, error) {
	socketName := fmt.Sprintf("%s-podman-api.sock", name)
	return socketDir.AppendToNewVMFile(socketName, nil)
}

func ignitionSocket(name string, socketDir *define.VMFile) (*define.VMFile, error) {
	socketName := fmt.Sprintf("%s-ignition.sock", name)
	return socketDir.AppendToNewVMFile(socketName, nil)
}
