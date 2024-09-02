package vmconfigs

import (
	"bauklotze/pkg/machine/machineDefine"
	"fmt"
)

func readySocket(name string, machineRuntimeDir *machineDefine.VMFile) (*machineDefine.VMFile, error) {
	socketName := name + ".sock"
	return machineRuntimeDir.AppendToNewVMFile(socketName)
}

func gvProxySocket(name string, machineRuntimeDir *machineDefine.VMFile) (*machineDefine.VMFile, error) {
	socketName := fmt.Sprintf("%s-gvproxy.sock", name)
	return machineRuntimeDir.AppendToNewVMFile(socketName)
}

func apiSocket(name string, socketDir *machineDefine.VMFile) (*machineDefine.VMFile, error) {
	socketName := name + "-api.sock"
	return socketDir.AppendToNewVMFile(socketName)
}
