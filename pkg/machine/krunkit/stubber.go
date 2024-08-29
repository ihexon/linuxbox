package krunkit

import (
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/apple"
	"bauklotze/pkg/machine/apple/hvhelper"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/utils"
	"fmt"
	vfConfig "github.com/crc-org/vfkit/pkg/config"
	"strconv"
)

type LibKrunStubber struct {
	vmconfigs.AppleKrunkitConfig
}

const (
	krunkitBinary = "krunkit"
	localhostURI  = "http://localhost"
)

func (l LibKrunStubber) CreateVM(opts define.CreateVMOpts, mc *vmconfigs.MachineConfig) error {
	mc.AppleKrunkitHypervisor = new(vmconfigs.AppleKrunkitConfig)
	mc.AppleKrunkitHypervisor.Krunkit = hvhelper.Helper{}
	bl := vfConfig.NewEFIBootloader(fmt.Sprintf("%s/efi-bl-%s", opts.Dirs.DataDir.GetPath(), opts.Name), true)
	mc.AppleKrunkitHypervisor.Krunkit.VirtualMachine = vfConfig.NewVirtualMachine(uint(mc.Resources.CPUs), uint64(mc.Resources.Memory), bl)
	randPort, err := utils.GetRandomPort()
	if err != nil {
		return err
	}

	mc.AppleKrunkitHypervisor.Krunkit.Endpoint = localhostURI + ":" + strconv.Itoa(randPort)
	virtiofsMounts := make([]machine.VirtIoFs, 0, len(mc.Mounts))
	for _, mnt := range mc.Mounts {
		virtiofsMounts = append(virtiofsMounts, machine.MountToVirtIOFs(mnt))
	}
	//virtIOIgnitionMounts, err := apple.GenerateSystemDFilesForVirtiofsMounts(virtiofsMounts)
	if err != nil {
		return err
	}
	return apple.ResizeDisk(mc, mc.Resources.DiskSize)
}
