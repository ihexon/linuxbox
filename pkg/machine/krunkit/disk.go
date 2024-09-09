package krunkit

import (
	"bauklotze/pkg/machine/machineDefine"
	"bauklotze/pkg/machine/vmconfigs"
	strongunits "bauklotze/pkg/storage"
	"errors"
	"fmt"
	vfConfig "github.com/crc-org/vfkit/pkg/config"
	"github.com/sirupsen/logrus"
	"os"
)

func ResizeDisk(mc *vmconfigs.MachineConfig, newSize strongunits.GiB) error {
	logrus.Debugf("resizing %s to %d bytes", mc.ImagePath.GetPath(), newSize.ToBytes())
	return os.Truncate(mc.ImagePath.GetPath(), int64(newSize.ToBytes()))
}

func VirtIOFsToVFKitVirtIODevice(mounts []*vmconfigs.Mount) ([]vfConfig.VirtioDevice, error) {
	virtioDevices := make([]vfConfig.VirtioDevice, 0, len(mounts))
	for _, vol := range mounts {
		virtfsDevice, err := vfConfig.VirtioFsNew(vol.Source, vol.Tag)
		if err != nil {
			return nil, err
		}
		virtioDevices = append(virtioDevices, virtfsDevice)
	}
	return virtioDevices, nil
}

func SetProviderAttrs(mc *vmconfigs.MachineConfig, opts machineDefine.SetOptions, state machineDefine.Status) error {
	if state != machineDefine.Stopped {
		return errors.New("unable to change settings unless vm is stopped")
	}

	if opts.DiskSize != nil {
		if err := ResizeDisk(mc, *opts.DiskSize); err != nil {
			return err
		}
	}

	// Not support for now
	//if opts.Rootful != nil && mc.HostUser.Rootful != *opts.Rootful {
	//	if err := mc.SetRootful(*opts.Rootful); err != nil {
	//		return err
	//	}
	//}

	// Not support for applehv
	if opts.USBs != nil {
		return fmt.Errorf("changing USBs not supported for applehv machines")
	}

	// VFKit does not require saving memory, disk, or cpu
	return nil
}
