//go:build !drawin && !linux && windows

package wsl

import (
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/machineDefine"
	"bauklotze/pkg/machine/vmconfigs"
	gvproxy "github.com/containers/gvisor-tap-vsock/pkg/types"
)

type WSLStubber struct {
	machineDefine.WSLConfig
}

func (w WSLStubber) GetDisk(userInputPath string, dirs *machineDefine.MachineDirs, mc *vmconfigs.MachineConfig) error {
	//TODO implement me
	panic("implement me")
}

func (w WSLStubber) CreateVM(opts machineDefine.CreateVMOpts, mc *vmconfigs.MachineConfig) error {
	//TODO implement me
	panic("implement me")
}

func (w WSLStubber) StopVM(mc *vmconfigs.MachineConfig, hardStop bool) error {
	//TODO implement me
	panic("implement me")
}

func (w WSLStubber) MountType() vmconfigs.VolumeMountType {
	//TODO implement me
	panic("implement me")
}

func (w WSLStubber) RequireExclusiveActive() bool {
	//TODO implement me
	panic("implement me")
}

func (w WSLStubber) State(mc *vmconfigs.MachineConfig, bypass bool) (machineDefine.Status, error) {
	//TODO implement me
	panic("implement me")
}

func (w WSLStubber) UpdateSSHPort(mc *vmconfigs.MachineConfig, port int) error {
	//TODO implement me
	panic("implement me")
}

func (w WSLStubber) UseProviderNetworkSetup() bool {
	//TODO implement me
	panic("implement me")
}

func (w WSLStubber) StartNetworking(mc *vmconfigs.MachineConfig, cmd *gvproxy.GvproxyCommand) error {
	//TODO implement me
	panic("implement me")
}

func (w WSLStubber) PostStartNetworking(mc *vmconfigs.MachineConfig, noInfo bool) error {
	//TODO implement me
	panic("implement me")
}

func (w WSLStubber) StartVM(mc *vmconfigs.MachineConfig) (func() error, func() error, error) {
	//TODO implement me
	panic("implement me")
}

func (w WSLStubber) MountVolumesToVM(mc *vmconfigs.MachineConfig, quiet bool) error {
	//TODO implement me
	panic("implement me")
}

func (w WSLStubber) Exists(name string) (bool, error) {
	return isWSLExist(env.WithBugBoxPrefix(name))
}

func (w WSLStubber) VMType() machineDefine.VMType {
	return machineDefine.WSLVirt
}

func (w WSLStubber) Remove(mc *vmconfigs.MachineConfig) ([]string, func() error, error) {

	wslRemoveFunc := func() error {
		if err := runCmdPassThrough(FindWSL(), "--unregister", env.WithBugBoxPrefix(mc.Name)); err != nil {
			return err
		}
		return nil
	}

	return []string{}, wslRemoveFunc, nil
}

func isRunning(name string) (bool, error) {
	dist := env.WithBugBoxPrefix(name)
	running, err := isWSLRunning(dist)
	if err != nil {
		return false, err
	}

	// TODO: isPodmanAPI Running
	//sysd := false
	//if wsl {
	//	sysd, err = isSystemdRunning(dist)
	//
	//	if err != nil {
	//		return false, err
	//	}
	// }
	return running, err
}
