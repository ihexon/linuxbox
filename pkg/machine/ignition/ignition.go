package ignition

import (
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/vmconfigs"
)

type DynamicIgnition struct {
	Commands []string
}

func NewDynamicIgnition() DynamicIgnition {
	return DynamicIgnition{
		Commands: []string{},
	}
}

func (ign *DynamicIgnition) GetDynamicIgnition() []string {
	return ign.Commands
}

func ServeIgnitionOverSock(mc *vmconfigs.MachineConfig) (*DynamicIgnition, error) {
	ign := NewDynamicIgnition()
	err := ign.generateSetTimeZoneIgnitionCfg()
	if err != nil {
		return nil, err
	}
	pub, err := machine.GetSSHKeys(mc.SSH.IdentityPath)
	if err != nil {
		return nil, err
	}

	err = ign.generateSSHIgnitionCfg(pub)
	if err != nil {
		return nil, err
	}

	err = ign.generateMountsIgnitionCfg(mc.Mounts)
	if err != nil {
		return nil, err
	}

	return &ign, err

}
