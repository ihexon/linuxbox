package vmconfigs

import (
	"bauklotze/pkg/machine/define"
	"fmt"
)

func gvProxySocket(name string, machineRuntimeDir *define.VMFile) (*define.VMFile, error) {
	socketName := fmt.Sprintf("%s-gvproxy.sock", name)
	return machineRuntimeDir.AppendToNewVMFile(socketName)
}
