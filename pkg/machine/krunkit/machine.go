//go:build darwin && arm64

package krunkit

import (
	"bauklotze/pkg/config"
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/ignition"
	"bauklotze/pkg/machine/sockets"
	"bauklotze/pkg/machine/vmconfigs"
	"context"
	"errors"
	"fmt"
	"github.com/containers/storage/pkg/fileutils"
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

	// Note: After Ignition, We send ready to `readySocket.GetPath()`
	readyDevice, err := vfConfig.VirtioVsockNew(1025, readySocket.GetPath(), true)
	if err != nil {
		return nil, nil, err
	}

	ignitionSocket, err := mc.IgnitionSocket()
	if err != nil {
		return nil, nil, err
	}

	// DO NOT CHANGE THE 1024 VSOCK PORT
	// See https://coreos.github.io/ignition/supported-platforms/
	ignitionDevice, err := vfConfig.VirtioVsockNew(1024, ignitionSocket.GetPath(), true)
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

// TODO, If there is an error,  it should return error
func readFileContent(path string) string {
	content, _ := os.ReadFile(path)

	return string(content)
}

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
	vm.Devices = append(vm.Devices, defaultDevices...)
	vm.Devices = append(vm.Devices, netDevice)

	// If the --external-disk flag is set, we need to create the disk **if it does not exist**
	if mc.ExternalDisk.GetPath() != "" {
		if err = fileutils.Exists(mc.ExternalDisk.GetPath()); err != nil {
			logrus.Warnf("external disk does not exist: %s", mc.ExternalDisk.GetPath())
			if err = CreateAndResizeDisk(mc.ExternalDisk.GetPath(), 500); err != nil {
				return nil, nil, err
			}
		}

		external_disk, err := vfConfig.VirtioBlkNew(mc.ExternalDisk.GetPath())
		if err != nil {
			return nil, nil, err
		}
		vm.Devices = append(vm.Devices, external_disk)
	}

	mounts, err := VirtIOFsToVFKitVirtIODevice(mc.Mounts)
	if err != nil {
		return nil, nil, err
	}
	vm.Devices = append(vm.Devices, mounts...)

	// To start the VM, we need to call krunkit
	cfg, err := config.Default()
	if err != nil {
		return nil, nil, err
	}

	cmdBinaryPath, err := cfg.FindHelperBinary(cmdBinary, true)
	if err != nil {
		return nil, nil, err
	}
	logrus.Infof("krunkit binary path is: %s", cmdBinaryPath)

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
	// Listen ready socket
	if err := readySocket.Delete(); err != nil {
		logrus.Warnf("unable to delete previous ready socket: %q", err)
	}
	readyListen, err := net.Listen("unix", readySocket.GetPath())
	if err != nil {
		return nil, nil, err
	} else {
		logrus.Infof("listening ready event on: %s", readySocket.GetPath())
	}
	// Wait for ready event coming...
	readyChan := make(chan error)
	go sockets.ListenAndWaitOnSocket(readyChan, readyListen)
	logrus.Debug("waiting for ready notification...")

	ignFile, err := mc.IgnitionFile()
	if err != nil {
		return nil, nil, err
	}

	ignBuilder := ignition.NewIgnitionBuilder(ignition.DynamicIgnitionV2{
		Name:      define.DefaultUserInGuest,
		Key:       readFileContent(mc.SSH.IdentityPath + ".pub"),
		TimeZone:  "local", // Auto detect timezone from locales
		VMType:    define.LibKrun,
		VMName:    define.DefaultMachineName,
		WritePath: ignFile.GetPath(),
		Rootful:   true,
	})

	err = ignBuilder.GenerateIgnitionConfig()
	if err != nil {
		return nil, nil, err
	}

	err = ignBuilder.Build()
	if err != nil {
		return nil, nil, err
	}

	ignSocket, err := mc.IgnitionSocket()
	if err != nil {
		return nil, nil, err
	}

	if err := ignSocket.Delete(); err != nil {
		logrus.Errorf("failed to delete the %s", ignSocket.GetPath())
		return nil, nil, err
	}

	logrus.Infof("Serving the ignition file over the socket: %s", ignSocket.GetPath())
	go func() {
		if err := ignition.ServeIgnitionOverSockV2(ignSocket, mc); err != nil {
			logrus.Errorf("failed to serve ignition file: %v", err)
			readyChan <- err
		}
	}()

	logrus.Infof("krunkit command-line: %v", cmd.Args)

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	returnFunc := func() error {
		processErrChan := make(chan error)
		machine.GlobalPIDs.SetKrunkitPID(cmd.Process.Pid)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Go routine to check if the process gvproxy and krunkit is running
		go func() {
			defer close(processErrChan)
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				if err := CheckProcessRunning("Krunkit", machine.GlobalPIDs.GetKrunkitPID()); err != nil {
					processErrChan <- err
					return
				}

				if err := CheckProcessRunning("Gvproxy", machine.GlobalPIDs.GetGvproxyPID()); err != nil {
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
			logrus.Infof("machine ready notification received")
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

func SetProviderAttrs(mc *vmconfigs.MachineConfig, opts define.SetOptions, state define.Status) error {
	if state != define.Stopped {
		return errors.New("unable to change settings unless vm is stopped")
	}
	// VFKit does not require saving memory, disk, or cpu
	return nil
}
