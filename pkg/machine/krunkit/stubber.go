//go:build darwin && arm64

package krunkit

import (
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/apple/hvhelper"
	"bauklotze/pkg/machine/machineDefine"
	"bauklotze/pkg/machine/shim/diskpull"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/utils"
	"fmt"
	gvproxy "github.com/containers/gvisor-tap-vsock/pkg/types"
	vfConfig "github.com/crc-org/vfkit/pkg/config"
	"strconv"
)

type LibKrunStubber struct {
	vmconfigs.AppleKrunkitConfig
}

func (l LibKrunStubber) State(mc *vmconfigs.MachineConfig, bypass bool) (machineDefine.Status, error) {
	return mc.AppleKrunkitHypervisor.Krunkit.State()
}

func (l LibKrunStubber) StopVM(mc *vmconfigs.MachineConfig, hardStop bool) error {
	return mc.AppleKrunkitHypervisor.Krunkit.Stop(hardStop, true)
}

func (l LibKrunStubber) GetDisk(userInputPath string, dirs *machineDefine.MachineDirs, mc *vmconfigs.MachineConfig) error {

	return diskpull.GetDisk(userInputPath, dirs, mc.ImagePath, l.VMType(), mc.Name)
}

func (l LibKrunStubber) Exists(name string) (bool, error) {
	// not applicable for libkrun (same as applehv)
	return false, nil
}

// MountVolumesToVM do nothing cause we write guest:/etc/fstab
func (l LibKrunStubber) MountVolumesToVM(mc *vmconfigs.MachineConfig, quiet bool) error {
	return nil
}

func (l LibKrunStubber) Remove(mc *vmconfigs.MachineConfig) ([]string, func() error, error) {
	return []string{}, func() error { return nil }, nil
}

func (l LibKrunStubber) RemoveAndCleanMachines(dirs *machineDefine.MachineDirs) error {
	return nil
}

func (l LibKrunStubber) StartNetworking(mc *vmconfigs.MachineConfig, cmd *gvproxy.GvproxyCommand) error {
	return StartGenericNetworking(mc, cmd)
}

func (l LibKrunStubber) PostStartNetworking(mc *vmconfigs.MachineConfig, noInfo bool) error {
	return nil
}

func (l LibKrunStubber) UserModeNetworkEnabled(mc *vmconfigs.MachineConfig) bool {
	return true
}

func (l LibKrunStubber) UseProviderNetworkSetup() bool {
	return false
}

func (l LibKrunStubber) RequireExclusiveActive() bool {
	return true
}

func (l LibKrunStubber) UpdateSSHPort(mc *vmconfigs.MachineConfig, port int) error {
	return nil
}

func (l LibKrunStubber) GetRosetta(mc *vmconfigs.MachineConfig) (bool, error) {
	return false, nil
}

func (l LibKrunStubber) MountType() vmconfigs.VolumeMountType {
	return vmconfigs.VirtIOFS
}

const (
	krunkitBinary = "krunkit"
	localhostURI  = "http://127.0.0.1"
)

func (l LibKrunStubber) CreateVM(opts machineDefine.CreateVMOpts, mc *vmconfigs.MachineConfig) error {
	mc.AppleKrunkitHypervisor = new(vmconfigs.AppleKrunkitConfig)
	mc.AppleKrunkitHypervisor.Krunkit = hvhelper.Helper{}
	bl := vfConfig.NewEFIBootloader(fmt.Sprintf("%s/efi-bl-%s", opts.Dirs.DataDir.GetPath(), opts.Name), true)
	mc.AppleKrunkitHypervisor.Krunkit.VirtualMachine = vfConfig.NewVirtualMachine(uint(mc.Resources.CPUs), uint64(mc.Resources.Memory), bl)
	randPort, err := utils.GetRandomPort()
	if err != nil {
		return err
	}
	// Endpoint is a string: http://127.0.0.1/[random_port]
	mc.AppleKrunkitHypervisor.Krunkit.Endpoint = localhostURI + ":" + strconv.Itoa(randPort)
	virtiofsMounts := make([]machine.VirtIoFs, 0, len(mc.Mounts))
	for _, mnt := range mc.Mounts {
		virtiofsMounts = append(virtiofsMounts, machine.MountToVirtIOFs(mnt))
	}
	//virtIOIgnitionMounts, err := apple.GenerateSystemDFilesForVirtiofsMounts(virtiofsMounts)
	if err != nil {
		return err
	}
	return ResizeDisk(mc, mc.Resources.DiskSize)
}

func (l LibKrunStubber) VMType() machineDefine.VMType {
	return machineDefine.LibKrun
}

func (l LibKrunStubber) StartVM(mc *vmconfigs.MachineConfig) (func() error, func() error, error) {
	bl := mc.AppleKrunkitHypervisor.Krunkit.VirtualMachine.Bootloader
	if bl == nil {
		return nil, nil, fmt.Errorf("unable to determine boot loader for this machine")
	}
	return StartGenericAppleVM(mc, krunkitBinary, bl, mc.AppleKrunkitHypervisor.Krunkit.Endpoint)
}
