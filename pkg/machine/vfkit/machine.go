//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package vfkit

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"bauklotze/pkg/libexec"
	"bauklotze/pkg/machine/events"

	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/ignition"
	"bauklotze/pkg/machine/sockets"
	"bauklotze/pkg/machine/vmconfigs"
	mypty "bauklotze/pkg/pty"

	vfConfig "github.com/crc-org/vfkit/pkg/config"
	"github.com/crc-org/vfkit/pkg/rest"
	"github.com/sirupsen/logrus"
)

func GetDefaultDevices(mc *vmconfigs.MachineConfig) ([]vfConfig.VirtioDevice, error) {
	var devices []vfConfig.VirtioDevice

	disk, err := vfConfig.VirtioBlkNew(mc.ImagePath.GetPath())
	if err != nil {
		return nil, fmt.Errorf("failed to create disk device: %w", err)
	}
	rng, err := vfConfig.VirtioRngNew()
	if err != nil {
		return nil, fmt.Errorf("failed to create rng device: %w", err)
	}

	externalDisk, err := vfConfig.VirtioBlkNew(mc.DataDisk.GetPath())
	if err != nil {
		return nil, fmt.Errorf("failed to create externalDisk device: %w", err)
	}

	serial, err := vfConfig.VirtioSerialNewStdio()
	if err != nil {
		return nil, fmt.Errorf("failed to create serial device: %w", err)
	}
	devices = append(devices, serial)

	devices = append(devices, disk, rng)
	// MUST APPEND AFTER disk PLEASE DO NOT CHANGE THE ORDER PLZ PLZ PLZ
	devices = append(devices, externalDisk)
	return devices, nil
}

func GetVfKitEndpointCMDArgs(endpoint string) ([]string, error) {
	if len(endpoint) == 0 {
		return nil, errors.New("endpoint cannot be empty")
	}
	restEndpoint, err := rest.NewEndpoint(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create endpoint: %w", err)
	}
	return restEndpoint.ToCmdLine() //nolint:wrapcheck
}

func StartGenericAppleVM(mc *vmconfigs.MachineConfig, cmdBinary string, bootloader vfConfig.Bootloader, endpoint string) (*exec.Cmd, func() error, error) {
	const applehvMACAddress = "5a:94:ef:e4:0c:ee"
	// Add networking
	netDevice, err := vfConfig.VirtioNetNew(applehvMACAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create net device: %w", err)
	}
	// Set user networking with gvproxy
	gvproxySocket, err := mc.GVProxySocket() // default-gvproxy.sock
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get gvproxy socket: %w", err)
	}

	// Before `netDevice.SetUnixSocketPath(gvproxySocket.GetPath())`, we need to wait on gvproxy to be running and aware,
	// There is a little chance that the gvproxy is not ready yet, so we need to wait for it.
	if err := sockets.WaitForSocketWithBackoffs(gvproxySocket.GetPath()); err != nil {
		return nil, nil, fmt.Errorf("failed to wait for gvproxy: %w", err)
	}

	netDevice.SetUnixSocketPath(gvproxySocket.GetPath())

	// create a one-time virtual machine for starting because we dont want all this information in the
	// machineconfig if possible.  the preference was to derive this stuff
	vm := vfConfig.NewVirtualMachine(uint(mc.Resources.CPUs), uint64(mc.Resources.Memory), bootloader)
	defaultDevices, err := GetDefaultDevices(mc)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get default devices: %w", err)
	}
	vm.Devices = append(vm.Devices, defaultDevices...)
	vm.Devices = append(vm.Devices, netDevice)

	mounts, err := VirtIOFsToVFKitVirtIODevice(mc.Mounts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert virtio fs to virtio device: %w", err)
	}
	vm.Devices = append(vm.Devices, mounts...)

	cmdBinaryPath, err := libexec.FindBinary(cmdBinary)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to find vfkit binary: %w", err)
	}
	logrus.Infof("krunkit binary path is: %s", cmdBinaryPath)

	vfkitCmd, err := vm.Cmd(cmdBinaryPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create krunkit command: %w", err)
	}

	// endpoint is krunkit rest api endpoint
	endpointArgs, err := GetVfKitEndpointCMDArgs(endpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get vfkit endpoint args: %w", err)
	}

	vfkitCmd.Args = append(vfkitCmd.Args, endpointArgs...)

	err = ignition.GenerateIgnScripts(mc)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate ignition scripts: %w", err)
	}

	logrus.Infof("vfkit command-line: %v", vfkitCmd.Args)

	// Run krunkit in pty, the pty should never close because the krunkit is a background process
	events.NotifyRun(events.StartVMProvider, "vfkit staring...")
	ptmx, err := mypty.RunInPty(vfkitCmd)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to run krunkit in pty: %w", err)
	}
	events.NotifyRun(events.StartVMProvider, "vfkit start finished")

	go func() {
		_, _ = io.Copy(os.Stdout, ptmx)
	}()
	machine.GlobalCmds.SetVMProviderCmd(vfkitCmd)

	mc.AppleVFkitHypervisor.Vfkit.BinaryPath, _ = define.NewMachineFile(cmdBinaryPath, nil)

	returnFunc := func() error {
		return nil
	}
	return vfkitCmd, returnFunc, nil
}
