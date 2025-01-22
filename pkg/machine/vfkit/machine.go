//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package vfkit

import (
	"bauklotze/pkg/machine/events"
	"bauklotze/pkg/machine/helper"
	"bauklotze/pkg/machine/ignition"
	"bauklotze/pkg/machine/vmconfig"
	mypty "bauklotze/pkg/pty"
	"fmt"
	"io"
	"os"

	vfConfig "github.com/crc-org/vfkit/pkg/config"
	"github.com/sirupsen/logrus"
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

func startVFKit(mc *vmconfig.MachineConfig) error {
	bootloader := mc.AppleVFkitHypervisor.Vfkit.VirtualMachine.Bootloader
	if bootloader == nil {
		return fmt.Errorf("unable to determine boot loader for this machine")
	}

	vmc := vfConfig.NewVirtualMachine(uint(mc.Resources.CPUs), uint64(mc.Resources.Memory), bootloader)

	defaultDevices, err := setupDevices(mc)
	if err != nil {
		return fmt.Errorf("failed to get default devices: %w", err)
	}
	vmc.Devices = append(vmc.Devices, defaultDevices...)

	vfkitBin := mc.Dirs.Hypervisor.Bin.GetPath()
	logrus.Infof("vfkit binary path is: %s", vfkitBin)

	vfkitCmd, err := vmc.Cmd(vfkitBin)
	if err != nil {
		return fmt.Errorf("failed to create vfkit command: %w", err)
	}

	err = ignition.GenerateIgnScripts(mc)
	if err != nil {
		return fmt.Errorf("failed to generate ignition scripts: %w", err)
	}

	logrus.Infof("FULL VFKIT CMDLINE:%q", vfkitCmd.Args)
	events.NotifyRun(events.StartVMProvider, "vfkit staring...")

	// Run vfkit in pty, the pty should never close because the vfkit is a background process
	ptmx, err := mypty.RunInPty(vfkitCmd)
	if err != nil {
		return fmt.Errorf("failed to run vfkit in pty: %w", err)
	}
	mc.VmmCmd = vfkitCmd
	go func() {
		_, _ = io.Copy(os.Stdout, ptmx)
	}()

	events.NotifyRun(events.StartVMProvider, "vfkit started")

	return nil
}
