package apple

import (
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/shim/diskpull"
	"bauklotze/pkg/machine/vmconfigs"
)

type AppleHVStubber struct {
	vmconfigs.AppleHVConfig
}

func (a AppleHVStubber) VMType() define.VMType {
	return define.AppleHvVirt
}

func (a AppleHVStubber) Exists(name string) (bool, error) {
	return false, nil
}

func (a AppleHVStubber) GetDisk(userInputPath string, dirs *define.MachineDirs, mc *vmconfigs.MachineConfig) error {
	return diskpull.GetDisk(userInputPath, dirs, mc.ImagePath, a.VMType(), mc.Name)

}

func (a AppleHVStubber) CreateVM(opts define.CreateVMOpts, mc *vmconfigs.MachineConfig) error {
	return nil
}

func (a *AppleHVStubber) StopVM(mc *vmconfigs.MachineConfig, _ bool) error {
	return mc.AppleHypervisor.Vfkit.Stop(false, true)
}
