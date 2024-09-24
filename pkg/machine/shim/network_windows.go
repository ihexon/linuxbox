package shim

import (
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfigs"
)

func setupMachineSockets(mc *vmconfigs.MachineConfig, dirs *define.MachineDirs) ([]string, string, machine.APIForwardingState, error) {
	return nil, "", 0, nil
}
