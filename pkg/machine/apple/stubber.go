package apple

import (
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/apple/hvhelper"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/shim/diskpull"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/network"
	"fmt"
	vfConfig "github.com/crc-org/vfkit/pkg/config"
	"strconv"
)

type AppleHVStubber struct {
	vmconfigs.AppleVfkitConfig
}

func (a AppleHVStubber) VMType() define.VMType {
	return define.AppleHvVirt
}

func (a AppleHVStubber) Exists(name string) (bool, error) {
	return false, nil
}

// GetDisk : status ok !
func (a AppleHVStubber) GetDisk(userInputPath string, dirs *define.MachineDirs, mc *vmconfigs.MachineConfig) error {
	return diskpull.GetDisk(userInputPath, dirs, mc.ImagePath, a.VMType(), mc.Name)
}

func (a AppleHVStubber) CreateVM(opts define.CreateVMOpts, mc *vmconfigs.MachineConfig) error {
	mc.AppleHypervisor.Vfkit = hvhelper.Helper{}
	bl := vfConfig.NewEFIBootloader(fmt.Sprintf("%s/efi-bl-%s", opts.Dirs.DataDir.GetPath(), opts.Name), true)
	mc.AppleHypervisor.Vfkit.VirtualMachine = vfConfig.NewVirtualMachine(
		uint(mc.Resources.CPUs),
		uint64(mc.Resources.Memory),
		bl)

	randPort, err := network.GetRandomPort()
	if err != nil {
		return err
	}
	mc.AppleHypervisor.Vfkit.Endpoint = localhostURI + ":" + strconv.Itoa(randPort)

	virtiofsMounts := make([]machine.VirtIoFs, 0, len(mc.Mounts))
	for _, mnt := range mc.Mounts {
		virtiofsMounts = append(virtiofsMounts, machine.MountToVirtIOFs(mnt))
	}
	return ResizeDisk(mc, mc.Resources.DiskSize)
}

func (a AppleHVStubber) MountType() vmconfigs.VolumeMountType {
	return vmconfigs.VirtIOFS
}

func (a *AppleHVStubber) StopVM(mc *vmconfigs.MachineConfig, _ bool) error {
	return mc.AppleHypervisor.Vfkit.Stop(false, true)
}
