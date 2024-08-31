package vmconfigs

import (
	"bauklotze/pkg/machine/machineDefine"
	"fmt"
)

func gvProxySocket(name string, machineRuntimeDir *machineDefine.VMFile) (*machineDefine.VMFile, error) {
	socketName := fmt.Sprintf("%s-gvproxy.sock", name)
	return machineRuntimeDir.AppendToNewVMFile(socketName)
}
