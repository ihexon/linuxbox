package vmconfigs

import (
	"bauklotze/pkg/machine/define"
)

func (mc *MachineConfig) ReadySocket() (*define.VMFile, error) {
	rtDir, err := mc.RuntimeDir()
	if err != nil {
		return nil, err
	}
	return readySocket(mc.Name, rtDir)
}

func (mc *MachineConfig) IgnitionSocket() (*define.VMFile, error) {
	rtDir, err := mc.RuntimeDir()
	if err != nil {
		return nil, err
	}
	return ignitionSocket(mc.Name, rtDir)
}
