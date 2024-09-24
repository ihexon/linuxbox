package vmconfigs

import (
	"bauklotze/pkg/machine/define"
	"fmt"
)

func readySocket(name string, machineRuntimeDir *define.VMFile) (*define.VMFile, error) {
	socketName := name + "-ready.sock"
	return machineRuntimeDir.AppendToNewVMFile(socketName)
}

func gvProxySocket(name string, machineRuntimeDir *define.VMFile) (*define.VMFile, error) {
	socketName := fmt.Sprintf("%s-gvproxy.sock", name)
	return machineRuntimeDir.AppendToNewVMFile(socketName)
}

func apiSocket(name string, socketDir *define.VMFile) (*define.VMFile, error) {
	socketName := name + "-api.sock"
	return socketDir.AppendToNewVMFile(socketName)
}

func ignitionSocket(name string, socketDir *define.VMFile) (*define.VMFile, error) {
	socketName := name + "-ignition.sock"
	return socketDir.AppendToNewVMFile(socketName)
}
