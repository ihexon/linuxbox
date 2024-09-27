package vmconfigs

import "bauklotze/pkg/machine/define"

func (mc *MachineConfig) LogFile() (*define.VMFile, error) {
	rtDir, err := mc.RuntimeDir()
	if err != nil {
		return nil, err
	}
	return rtDir.AppendToNewVMFile(mc.Name+".log", nil)
}
