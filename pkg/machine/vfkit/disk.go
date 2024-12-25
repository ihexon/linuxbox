//go:build darwin

package vfkit

import (
	"fmt"

	"bauklotze/pkg/machine/vmconfigs"

	vfConfig "github.com/crc-org/vfkit/pkg/config"
)

func VirtIOFsToVFKitVirtIODevice(mounts []*vmconfigs.Mount) ([]vfConfig.VirtioDevice, error) {
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
