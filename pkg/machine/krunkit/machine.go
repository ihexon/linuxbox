//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package krunkit

import (
	"fmt"

	"bauklotze/pkg/machine/helper"
	"bauklotze/pkg/machine/vmconfig"

	vfConfig "github.com/crc-org/vfkit/pkg/config"
)

const applehvMACAddress = "5a:94:ef:e4:0c:ee"

// setupKrunkitDevices add devices into VirtualMachine
func setupDevices(mc *vmconfig.MachineConfig) ([]vfConfig.VirtioDevice, error) {
	var devices []vfConfig.VirtioDevice

	disk, err := vfConfig.VirtioBlkNew(mc.Bootable.Image.GetPath())
	if err != nil {
		return nil, fmt.Errorf("failed to create bootable disk device: %w", err)
	}
	rng, err := vfConfig.VirtioRngNew()
	if err != nil {
		return nil, fmt.Errorf("failed to create rng device: %w", err)
	}

	// externalDisk is the disk used to store the user data, it will format as ext4
	externalDisk, err := vfConfig.VirtioBlkNew(mc.DataDisk.Image.GetPath())
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

	// externalDisk **must** be at the end of the device
	devices = append(devices, disk, rng, netDevice, externalDisk)

	VirtIOMounts, err := helper.VirtIOFsToVFKitVirtIODevice(mc.Mounts)
	if err != nil {
		return nil, fmt.Errorf("failed to convert virtio fs to virtio device: %w", err)
	}
	devices = append(devices, VirtIOMounts...)

	return devices, nil
}
