package vmconfigs

import "bauklotze/pkg/machine/machineDefine"

func (mc *MachineConfig) LogFile() (*machineDefine.VMFile, error) {
	rtDir, err := mc.RuntimeDir()
	if err != nil {
		return nil, err
	}
	return rtDir.AppendToNewVMFile(mc.Name + ".log")
}
