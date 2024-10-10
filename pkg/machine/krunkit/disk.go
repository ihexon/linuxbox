//go:build darwin && arm64

package krunkit

import (
	"bauklotze/pkg/machine/vmconfigs"
	"github.com/containers/common/pkg/strongunits"
	vfConfig "github.com/crc-org/vfkit/pkg/config"
	"github.com/sirupsen/logrus"
	"os"
)

func CreateAndResizeDisk(diskPath string, newSize strongunits.GiB) error {
	file, err := os.Create(diskPath)
	if err != nil {
		return err
	}
	defer file.Close()
	if err = os.Truncate(diskPath, int64(newSize.ToBytes())); err != nil {
		return err
	}
	return nil
}

// ResizeDisk Dangerous This function is not safe to use in production, because the raw disk is being truncated.
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
