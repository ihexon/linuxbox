package helper

import (
	"bauklotze/pkg/machine/io"
	"bauklotze/pkg/machine/volumes"
	"fmt"
	"os"

	"github.com/containers/common/pkg/strongunits"
	vfConfig "github.com/crc-org/vfkit/pkg/config"
)

func VirtIOFsToVFKitVirtIODevice(mounts []*volumes.Mount) ([]vfConfig.VirtioDevice, error) {
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

func CreateAndResizeDisk(f *io.FileWrapper, newSize strongunits.GiB) error {
	if f.Exist() {
		if err := f.Delete(true); err != nil {
			return fmt.Errorf("failed to delete disk: %w", err)
		}
	}

	if err := f.MakeBaseDir(); err != nil {
		return fmt.Errorf("failed to make base dir: %w", err)
	}

	file, err := os.Create(f.GetPath())
	if err != nil {
		return fmt.Errorf("failed to create disk: %s, %w", f.GetPath(), err)
	}
	defer file.Close()
	if err = os.Truncate(f.GetPath(), int64(newSize.ToBytes())); err != nil {
		return fmt.Errorf("failed to truncate disk: %w", err)
	}
	return nil
}
