package helper

import (
	"fmt"
	"os"

	"bauklotze/pkg/machine/volumes"

	"github.com/containers/common/pkg/strongunits"
	vfConfig "github.com/crc-org/vfkit/pkg/config"
)

func VirtIOFsToVFKitVirtIODevice(mounts []volumes.Mount) ([]vfConfig.VirtioDevice, error) {
	virtioDevices := make([]vfConfig.VirtioDevice, 0, len(mounts))
	for _, vol := range mounts {
		virtfsDevice, err := vfConfig.VirtioFsNew(vol.Source, vol.Tag)
		if err != nil {
			return nil, fmt.Errorf("failed to create virtio fs device: %w", err)
		}
		virtioDevices = append(virtioDevices, virtfsDevice)
	}
	return virtioDevices, nil
}

// CreateAndResizeDisk create a disk file with sizeInGB, and truncate it to sizeInGB.
func CreateAndResizeDisk(f string, sizeInGB int64) error {
	file, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create disk: %q, %w", f, err)
	}
	defer file.Close() //nolint:errcheck

	if err = os.Truncate(f, int64(strongunits.GiB(sizeInGB).ToBytes())); err != nil {
		return fmt.Errorf("failed to truncate disk: %w", err)
	}

	return nil
}
