//go:build darwin && arm64

package krunkit

import (
	"bauklotze/pkg/config"
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/sockets"
	"bauklotze/pkg/machine/vmconfigs"
	"context"
	"errors"
	"fmt"
	vfConfig "github.com/crc-org/vfkit/pkg/config"
	"github.com/crc-org/vfkit/pkg/rest"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"syscall"
	"time"
)

func GetDefaultDevices(mc *vmconfigs.MachineConfig) ([]vfConfig.VirtioDevice, *define.VMFile, error) {
	var devices []vfConfig.VirtioDevice

	disk, err := vfConfig.VirtioBlkNew(mc.ImagePath.GetPath())
	if err != nil {
		return nil, nil, err
	}
	rng, err := vfConfig.VirtioRngNew()
	if err != nil {
		return nil, nil, err
	}

	logfile, err := mc.LogFile()
	if err != nil {
		return nil, nil, err
	}
	serial, err := vfConfig.VirtioSerialNew(logfile.GetPath())
	if err != nil {
		return nil, nil, err
	}

	readySocket, err := mc.ReadySocket()
	if err != nil {
		return nil, nil, err
	}

	// Note: the connection from guest to host
	readyDevice, err := vfConfig.VirtioVsockNew(1025, readySocket.GetPath(), true)
	if err != nil {
		return nil, nil, err
	}

	ignitionSocket, err := mc.IgnitionSocket()
	if err != nil {
		return nil, nil, err
	}

	ignitionDevice, err := vfConfig.VirtioVsockNew(1026, ignitionSocket.GetPath(), true)
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
		return nil, err
	}
	return restEndpoint.ToCmdLine()
}

const applehvMACAddress = "5a:94:ef:e4:0c:ee"

var (
	gvProxyWaitBackoff        = 500 * time.Millisecond
	gvProxyMaxBackoffAttempts = 6
)

func StartGenericAppleVM(mc *vmconfigs.MachineConfig, cmdBinary string, bootloader vfConfig.Bootloader, endpoint string) (func() error, func() error, error) {
	// Add networking
	netDevice, err := vfConfig.VirtioNetNew(applehvMACAddress)
	if err != nil {
		return nil, nil, err
	}
	// Set user networking with gvproxy
	gvproxySocket, err := mc.GVProxySocket()
	if err != nil {
		return nil, nil, err
	}

	// Wait on gvproxy to be running and aware
	if err := sockets.WaitForSocketWithBackoffs(gvProxyMaxBackoffAttempts, gvProxyWaitBackoff, gvproxySocket.GetPath(), "gvproxy"); err != nil {
		return nil, nil, err
	}

	netDevice.SetUnixSocketPath(gvproxySocket.GetPath())

	// create a one-time virtual machine for starting because we dont want all this information in the
	// machineconfig if possible.  the preference was to derive this stuff
	vm := vfConfig.NewVirtualMachine(uint(mc.Resources.CPUs), uint64(mc.Resources.Memory), bootloader)

	defaultDevices, readySocket, err := GetDefaultDevices(mc)
	if err != nil {
		return nil, nil, err
	}

	vm.Devices = append(vm.Devices, defaultDevices...)
	vm.Devices = append(vm.Devices, netDevice)

	mounts, err := VirtIOFsToVFKitVirtIODevice(mc.Mounts)
	if err != nil {
		return nil, nil, err
	}
	vm.Devices = append(vm.Devices, mounts...)

	// To start the VM, we need to call vfkit
	cfg, err := config.Default()
	if err != nil {
		return nil, nil, err
	}

	cmdBinaryPath, err := cfg.FindHelperBinary(cmdBinary, true)
	if err != nil {
		return nil, nil, err
	}

	logrus.Debugf("helper binary path is: %s", cmdBinaryPath)

	cmd, err := vm.Cmd(cmdBinaryPath)
	if err != nil {
		return nil, nil, err
	}

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	// endpoint is krunkit rest api endpoint
	endpointArgs, err := GetVfKitEndpointCMDArgs(endpoint)
	if err != nil {
		return nil, nil, err
	}

	cmd.Args = append(cmd.Args, endpointArgs...)

	// TODO: firset boot things
	firstBoot, err := mc.IsFirstBoot()
	if err != nil {
		return nil, nil, err
	}
	if firstBoot {
	}

	logrus.Infof("listening for ready on: %s", readySocket.GetPath())
	if err := readySocket.Delete(); err != nil {
		logrus.Warnf("unable to delete previous ready socket: %q", err)
	}

	readyListen, err := net.Listen("unix", readySocket.GetPath())
	if err != nil {
		return nil, nil, err
	}

	logrus.Debug("waiting for ready notification")
	readyChan := make(chan error)
	go sockets.ListenAndWaitOnSocket(readyChan, readyListen)

	logrus.Debugf("helper command-line: %v", cmd.Args)

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	ignitionSocket, err := mc.IgnitionSocket()
	if err != nil {
		return nil, nil, err
	}
	logrus.Infof("IgnitionSocket on: %s", ignitionSocket.GetPath())
	if err = ignitionSocket.Delete(); err != nil {
		logrus.Warnf("unable to delete previous ready socket: %q", err)
	}
	ignListen, err := net.Listen("unix", ignitionSocket.GetPath())

	if err != nil {
		return nil, nil, err
	}

	err = sockets.ListenAndExecCommandOnUnixSocketFile(ignListen, mc)
	if err != nil {
		return nil, nil, err
	}
	logrus.Infof("Ignition finished")

	returnFunc := func() error {
		processErrChan := make(chan error)
		machine.GlobalPIDs.SetKrunkitPID(cmd.Process.Pid)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			defer close(processErrChan)
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				if err := CheckProcessRunning(cmdBinary, machine.GlobalPIDs.GetKrunkitPID()); err != nil {
					processErrChan <- err
					return
				}
				// lets poll status every half second
				time.Sleep(500 * time.Millisecond)
			}
		}()

		// wait for either socket or to be ready or process to have exited
		select {
		case err := <-processErrChan:
			if err != nil {
				return err
			}
		case err := <-readyChan:
			if err != nil {
				return err
			}
			logrus.Infof("podman ready notification received")
		}
		return nil
	}
	return cmd.Process.Release, returnFunc, nil
}

// CheckProcessRunning checks non blocking if the pid exited
// returns nil if process is running otherwise an error if not
func CheckProcessRunning(processName string, pid int) error {
	var status syscall.WaitStatus
	pid, err := syscall.Wait4(pid, &status, syscall.WNOHANG, nil)
	if err != nil {
		return fmt.Errorf("failed to read %s process status: %w", processName, err)
	}
	if pid > 0 {
		// child exited
		return fmt.Errorf("%s exited unexpectedly with exit code %d", processName, status.ExitStatus())
	}
	return nil
}
