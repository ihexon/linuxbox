//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package vfkit

import (
	"fmt"

	"bauklotze/pkg/machine/helper"
	"bauklotze/pkg/machine/vmconfig"

	vfConfig "github.com/crc-org/vfkit/pkg/config"
)

const applehvMACAddress = "5a:94:ef:e4:0c:ee"

func setupDevices(mc *vmconfig.MachineConfig) ([]vfConfig.VirtioDevice, error) {
	var devices []vfConfig.VirtioDevice
	bootableDisk, err := vfConfig.VirtioBlkNew(mc.Bootable.Image.GetPath())
	if err != nil {
		return nil, fmt.Errorf("failed to create bootable disk device: %w", err)
	}

	rng, err := vfConfig.VirtioRngNew()
	if err != nil {
		return nil, fmt.Errorf("failed to create rng device: %w", err)
	}

	// dataDisk is the disk used to store the user data, it will format as ext4
	dataDisk, err := vfConfig.VirtioBlkNew(mc.DataDisk.Image.GetPath())
	if err != nil {
		return nil, fmt.Errorf("failed to create externalDisk device: %w", err)
	}

	// using gvproxy as network backend
	netDevice, err := vfConfig.VirtioNetNew(applehvMACAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create net device: %w", err)
	}
	gvproxySocket, err := mc.GVProxyNetworkBackendSocks()
	if err != nil {
		return nil, fmt.Errorf("failed to get gvproxy socket: %w", err)
	}
	netDevice.SetUnixSocketPath(gvproxySocket.GetPath())

	// dataDisk **must** behind at the bootableDisk
	devices = append(devices, bootableDisk, rng, netDevice, dataDisk)

	// Add VirtIOFS devices
	VirtIOMounts, err := helper.VirtIOFsToVFKitVirtIODevice(mc.Mounts)
	if err != nil {
		return nil, fmt.Errorf("failed to convert virtio fs to virtio device: %w", err)
	}
	devices = append(devices, VirtIOMounts...)

	// Add serial console which vfkit log message into pty console
	serial, err := vfConfig.VirtioSerialNewStdio()
	if err != nil {
		return nil, fmt.Errorf("failed to create serial device: %w", err)
	}
	devices = append(devices, serial)

	return devices, nil
}
