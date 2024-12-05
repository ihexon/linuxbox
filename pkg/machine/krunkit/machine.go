//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin && arm64

package krunkit

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	"bauklotze/pkg/config"
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/ignition"
	"bauklotze/pkg/machine/sockets"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/system"

	"github.com/containers/storage/pkg/fileutils"
	vfConfig "github.com/crc-org/vfkit/pkg/config"
	"github.com/crc-org/vfkit/pkg/rest"
	"github.com/sirupsen/logrus"
)

func GetDefaultDevices(mc *vmconfigs.MachineConfig) ([]vfConfig.VirtioDevice, *define.VMFile, error) {
	var devices []vfConfig.VirtioDevice

	disk, err := vfConfig.VirtioBlkNew(mc.ImagePath.GetPath())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create disk device: %w", err)
	}
	rng, err := vfConfig.VirtioRngNew()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create rng device: %w", err)
	}

	logfile, err := mc.LogFile()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get log file: %w", err)
	}
	serial, err := vfConfig.VirtioSerialNew(logfile.GetPath())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create serial device: %w", err)
	}

	readySocket, err := mc.ReadySocket()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get ready socket: %w", err)
	}

	// Note: After Ignition, We send ready to `readySocket.GetPath()`
	readyDevice, err := vfConfig.VirtioVsockNew(1025, readySocket.GetPath(), true) //nolint:mnd

	if err != nil {
		return nil, nil, fmt.Errorf("failed to create ready device: %w", err)
	}

	ignitionSocket, err := mc.IgnitionSocket()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get ignition socket: %w", err)
	}

	// DO NOT CHANGE THE 1024 VSOCK PORT
	// See https://coreos.github.io/ignition/supported-platforms/
	ignitionDevice, err := vfConfig.VirtioVsockNew(1024, ignitionSocket.GetPath(), true) //nolint:mnd
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create ignition device: %w", err)
	}
	devices = append(devices, disk, rng, readyDevice, ignitionDevice)

	if mc.AppleKrunkitHypervisor == nil || !logrus.IsLevelEnabled(logrus.DebugLevel) {
		// If libkrun is the provider and we want to show the debug console,
		// don't add a virtio serial device to avoid redirecting the output.
		devices = append(devices, serial)
	}

	return devices, readySocket, nil
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

var (
	gvProxyWaitBackoff        = 100 * time.Millisecond
	gvProxyMaxBackoffAttempts = 6
)

// TODO, If there is an error,  it should return error
func readFileContent(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		logrus.Errorf("failed to read sshkey.pub content: %s", path)
		return ""
	}
	return string(content)
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
	if err := sockets.WaitForSocketWithBackoffs(gvProxyMaxBackoffAttempts, gvProxyWaitBackoff, gvproxySocket.GetPath(), "gvproxy"); err != nil {
		return nil, nil, fmt.Errorf("failed to wait for gvproxy: %w", err)
	}

	netDevice.SetUnixSocketPath(gvproxySocket.GetPath())

	// create a one-time virtual machine for starting because we dont want all this information in the
	// machineconfig if possible.  the preference was to derive this stuff
	vm := vfConfig.NewVirtualMachine(uint(mc.Resources.CPUs), uint64(mc.Resources.Memory), bootloader)
	defaultDevices, readySocket, err := GetDefaultDevices(mc)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get default devices: %w", err)
	}
	vm.Devices = append(vm.Devices, defaultDevices...)
	vm.Devices = append(vm.Devices, netDevice)

	if mc.DataDisk.GetPath() != "" {
		if err = fileutils.Exists(mc.DataDisk.GetPath()); err != nil {
			logrus.Warnf("external disk does not exist: %s", mc.DataDisk.GetPath())
			if err = system.CreateAndResizeDisk(mc.DataDisk.GetPath(), 100); err != nil { //nolint:mnd
				return nil, nil, fmt.Errorf("failed to create and resize disk: %w", err)
			}
		}

		externalDisk, err := vfConfig.VirtioBlkNew(mc.DataDisk.GetPath())
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create external disk: %w", err)
		}
		vm.Devices = append(vm.Devices, externalDisk)
	}

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

	krunCmd.Stdout = os.Stdout
	krunCmd.Stderr = os.Stderr

	// endpoint is krunkit rest api endpoint
	endpointArgs, err := GetVfKitEndpointCMDArgs(endpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get vfkit endpoint args: %w", err)
	}

	krunCmd.Args = append(krunCmd.Args, endpointArgs...)
	// Add the "krun-log-level" flag for setting up the desired log level for libkrun's debug facilities.
	// Log level for libkrun (0=off, 1=error, 2=warn, 3=info, 4=debug, 5 or higher=trace)
	krunCmd.Args = append(krunCmd.Args, "--krun-log-level", "3")

	// Listen ready socket
	if err := readySocket.Delete(); err != nil {
		logrus.Warnf("unable to delete previous ready socket: %q", err)
		return nil, nil, fmt.Errorf("failed to delete previous ready socket: %w", err)
	}

	readyListen, err := net.Listen("unix", readySocket.GetPath())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen ready event: %w", err)
	} else {
		logrus.Infof("Listening ready event on: %s", readySocket.GetPath())
	}
	// Wait for ready event coming...
	readyChan := make(chan error)
	go sockets.ListenAndWaitOnSocket(readyChan, readyListen)
	logrus.Infof("Waiting for ready notification...")

	ignFile, err := mc.IgnitionFile()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get ignition file: %w", err)
	}

	ignBuilder := ignition.NewIgnitionBuilder(ignition.DynamicIgnitionV2{
		Name:           define.DefaultUserInGuest,
		Key:            readFileContent(mc.SSH.IdentityPath + ".pub"),
		TimeZone:       "local", // Auto detect timezone from locales
		VMType:         define.LibKrun,
		VMName:         mc.Name,
		WritePath:      ignFile.GetPath(),
		Rootful:        true,
		MachineConfigs: mc,
		UID:            0,
	})

	err = ignBuilder.GenerateIgnitionConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate ignition config: %w", err)
	}

	err = ignBuilder.Build()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build ignition file: %w", err)
	}

	ignSocket, err := mc.IgnitionSocket()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get ignition socket: %w", err)
	}

	if err := ignSocket.Delete(); err != nil {
		logrus.Errorf("failed to delete the %s", ignSocket.GetPath())
		return nil, nil, fmt.Errorf("failed to delete the %s: %w", ignSocket.GetPath(), err)
	}

	logrus.Infof("Serving the ignition file over the socket: %s", ignSocket.GetPath())

	go func() {
		if err := ignition.ServeIgnitionOverSockV2(ignSocket, mc); err != nil {
			logrus.Errorf("failed to serve ignition file: %v", err)
			readyChan <- err
		}
	}()

	logrus.Infof("krunkit command-line: %v", krunCmd.Args)

	if err := krunCmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start krunkit: %w", err)
	} else {
		machine.GlobalCmds.SetKrunCmd(krunCmd)
	}

	mc.AppleKrunkitHypervisor.Krunkit.BinaryPath, _ = define.NewMachineFile(cmdBinaryPath, nil)

	returnFunc := func() error {
		// wait for either socket or to be ready or process to have exited
		if err := <-readyChan; err != nil {
			return err
		}
		logrus.Infof("machine ready notification received")
		return nil
	}
	return krunCmd, returnFunc, nil
}
