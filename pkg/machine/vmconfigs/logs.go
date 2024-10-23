package vmconfigs

import "bauklotze/pkg/machine/define"

func (mc *MachineConfig) LogFile() (*define.VMFile, error) {
	logsDir, err := mc.LogsDir()
	if err != nil {
		return nil, err
	}
	return logsDir.AppendToNewVMFile(mc.Name+".log", nil)
}
