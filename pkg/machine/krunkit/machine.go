//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package krunkit

import (
	"bauklotze/pkg/machine/events"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"bauklotze/pkg/config"
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

	devices = append(devices, disk, rng)
	// MUST APPEND AFTER disk PLEASE DO NOT CHANGE THE ORDER PLZ PLZ PLZ
	devices = append(devices, externalDisk)

	return devices, nil
}

// GetVfKitEndpointCMDArgs converts the vfkit endpoint to a cmdline format
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

	// To start the VM, we need to call krunkit
	cfg := config.Default()

	cmdBinaryPath, err := cfg.FindHelperBinary(cmdBinary)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to find krunkit binary: %w", err)
	}
	logrus.Infof("krunkit binary path is: %s", cmdBinaryPath)

	krunCmd, err := vm.Cmd(cmdBinaryPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create krunkit command: %w", err)
	}

	// endpoint is krunkit rest api endpoint
	endpointArgs, err := GetVfKitEndpointCMDArgs(endpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get vfkit endpoint args: %w", err)
	}

	krunCmd.Args = append(krunCmd.Args, endpointArgs...)
	// Add the "krun-log-level" flag for setting up the desired log level for libkrun's debug facilities.
	// Log level for libkrun (0=off, 1=error, 2=warn, 3=info, 4=debug, 5 or higher=trace)
	krunCmd.Args = append(krunCmd.Args, "--krun-log-level", "3")

	err = ignition.GenerateIgnScripts(mc)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate ignition scripts: %w", err)
	}

	logrus.Infof("krunkit command-line: %v", krunCmd.Args)
	events.NotifyRun(events.StartVMProvider, "krunkit staring...")
	// Run krunkit in pty, the pty should never close because the krunkit is a background process
	ptmx, err := mypty.RunInPty(krunCmd)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to run krunkit in pty: %w", err)
	}
	go func() {
		_, _ = io.Copy(os.Stdout, ptmx)
	}()
	events.NotifyRun(events.StartVMProvider, "krunkit start done")

	machine.GlobalCmds.SetVMProviderCmd(krunCmd)

	mc.AppleKrunkitHypervisor.Krunkit.BinaryPath, _ = define.NewMachineFile(cmdBinaryPath, nil)

	returnFunc := func() error {
		return nil
	}
	return krunCmd, returnFunc, nil
}
