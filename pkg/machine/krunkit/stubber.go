//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package krunkit

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"

	"bauklotze/pkg/decompress"
	"bauklotze/pkg/machine/defconfig"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/events"
	"bauklotze/pkg/machine/ignition"
	"bauklotze/pkg/machine/vmconfig"
	"bauklotze/pkg/machine/volumes"
	"bauklotze/pkg/port"
	mypty "bauklotze/pkg/pty"

	gvproxy "github.com/containers/gvisor-tap-vsock/pkg/types"
	vfConfig "github.com/crc-org/vfkit/pkg/config"
	"github.com/sirupsen/logrus"
)

type LibKrunStubber struct {
	vmconfig.AppleKrunkitConfig
}

// ExtractBootable only support zstd compressed bootable image
func (l LibKrunStubber) ExtractBootable(userInputPath string, mc *vmconfig.MachineConfig) error {
	destDir := mc.Bootable.Image.GetPath()
	logrus.Infof("Try to decompress %s to %s", userInputPath, destDir)
	if err := decompress.Zstd(userInputPath, mc.Bootable.Image.GetPath()); err != nil {
		errors := fmt.Errorf("could not decompress %s to %s, %w", userInputPath, destDir, err)
		return errors
	}
	return nil
}

func (l LibKrunStubber) SetupProviderNetworking(mc *vmconfig.MachineConfig, gvcmd *gvproxy.GvproxyCommand) error {
	gvpNetworkBackend, err := mc.GVProxyNetworkBackendSocks()
	if err != nil {
		return fmt.Errorf("failed to get gvproxy networking backend socket: %w", err)
	}
	// make sure it does not exist before gvproxy is called
	if err := gvpNetworkBackend.Delete(true); err != nil {
		return fmt.Errorf("failed to delete gvproxy socket: %w", err)
	}
	gvcmd.AddVfkitSocket(fmt.Sprintf("unixgram://%s", gvpNetworkBackend.GetPath()))
	return nil
}

func (l LibKrunStubber) MountType() volumes.VolumeMountType {
	return volumes.VirtIOFS
}

func (l LibKrunStubber) CreateVMConfig(mc *vmconfig.MachineConfig) error {
	mc.AppleKrunkitHypervisor = new(vmconfig.AppleKrunkitConfig)
	mc.AppleKrunkitHypervisor.Krunkit = vmconfig.Helper{}
	bl := vfConfig.NewEFIBootloader(fmt.Sprintf("%s/efi-bl-%s", mc.Dirs.DataDir.GetPath(), mc.VMName), true)
	mc.AppleKrunkitHypervisor.Krunkit.VirtualMachine = vfConfig.NewVirtualMachine(uint(mc.Resources.CPUs), uint64(mc.Resources.Memory), bl)
	randPort, err := port.GetFree(0)
	if err != nil {
		return fmt.Errorf("failed to get random port: %w", err)
	}
	// Endpoint is a string: http://127.0.0.1/[random_port]
	mc.AppleKrunkitHypervisor.Krunkit.Endpoint = fmt.Sprintf("%s:%s", define.LocalHostURL, strconv.Itoa(randPort))
	mc.AppleKrunkitHypervisor.Krunkit.LogLevel = logrus.InfoLevel

	return nil
}

func (l LibKrunStubber) VMType() defconfig.VMType {
	return defconfig.LibKrun
}

func (l LibKrunStubber) StartVM(ctx context.Context, mc *vmconfig.MachineConfig) error {
	bootloader := mc.AppleKrunkitHypervisor.Krunkit.VirtualMachine.Bootloader
	if bootloader == nil {
		return fmt.Errorf("unable to determine boot loader for this machine")
	}

	vmc := vfConfig.NewVirtualMachine(uint(mc.Resources.CPUs), uint64(mc.Resources.Memory), bootloader)

	defaultDevices, err := setupDevices(mc)
	if err != nil {
		return fmt.Errorf("failed to get default devices: %w", err)
	}
	vmc.Devices = append(vmc.Devices, defaultDevices...)

	krunkitBin := mc.Dirs.Hypervisor.Bin.GetPath()
	logrus.Infof("krunkit binary path is: %s", krunkitBin)

	cmd, err := vmc.Cmd(krunkitBin)
	if err != nil {
		return fmt.Errorf("failed to create krunkit command: %w", err)
	}
	libsDir := mc.Dirs.Hypervisor.LibsDir.GetPath()

	// Add the "krun-log-level" allflag for setting up the desired log level for libkrun's debug facilities.
	// Log level for libkrun (0=off, 1=error, 2=warn, 3=info, 4=debug, 5 or higher=trace)
	cmd.Args = append(cmd.Args, "--krun-log-level", "3")

	myKrunKitCmd := exec.CommandContext(ctx, krunkitBin, cmd.Args[1:]...)
	myKrunKitCmd.Env = append(myKrunKitCmd.Env, fmt.Sprintf("DYLD_LIBRARY_PATH=%s", libsDir))

	if err = ignition.GenerateIgnScripts(mc); err != nil {
		return fmt.Errorf("failed to generate ignition scripts: %w", err)
	}

	logrus.Infof("FULL KRUNKIT CMDLINE: %q", myKrunKitCmd.Args)
	events.NotifyRun(events.StartVMProvider, "krunkit staring...")

	// Run krunkit in pty, the pty should never close because the krunkit is a background process
	ptmx, err := mypty.RunInPty(myKrunKitCmd)
	if err != nil {
		return fmt.Errorf("failed to run krunkit in pty: %w", err)
	}
	mc.VmmCmd = myKrunKitCmd

	go func() {
		_, _ = io.Copy(os.Stdout, ptmx)
	}()

	events.NotifyRun(events.StartVMProvider, "krunkit started")

	return nil
}
