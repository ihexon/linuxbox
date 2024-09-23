package vmconfigs

import (
	"bauklotze/pkg/machine/machineDefine"
)

func (mc *MachineConfig) ReadySocket() (*machineDefine.VMFile, error) {
	rtDir, err := mc.RuntimeDir()
	if err != nil {
		return nil, err
	}
	return readySocket(mc.Name, rtDir)
}

func (mc *MachineConfig) IgnitionSocket() (*machineDefine.VMFile, error) {
	rtDir, err := mc.RuntimeDir()
	if err != nil {
		return nil, err
	}
	return ignitionSocket(mc.Name, rtDir)
}
