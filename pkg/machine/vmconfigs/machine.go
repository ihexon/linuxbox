package vmconfigs

import "bauklotze/pkg/machine/define"

func (mc *MachineConfig) GVProxySocket() (*define.VMFile, error) {
	machineRuntimeDir, err := mc.RuntimeDir()
	if err != nil {
		return nil, err
	}
	return gvProxySocket(mc.Name, machineRuntimeDir)
}
