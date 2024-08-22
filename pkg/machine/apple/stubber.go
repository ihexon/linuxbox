package apple

import (
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfigs"
)

type AppleHVStubber struct {
	vmconfigs.AppleHVConfig
}

func (a AppleHVStubber) VMType() define.VMType {
	return define.AppleHvVirt
}

func (H AppleHVStubber) Exists(name string) (bool, error) {
	return false, nil
}

func (H AppleHVStubber) GetDisk(userInputPath string, mc *vmconfigs.MachineConfig) error {

}

func (H AppleHVStubber) CreateVM(opts define.CreateVMOpts, mc *vmconfigs.MachineConfig) error {

}

func (a *AppleHVStubber) StopVM(mc *vmconfigs.MachineConfig, _ bool) error {
	return mc.AppleHypervisor.Vfkit.Stop(false, true)
}
