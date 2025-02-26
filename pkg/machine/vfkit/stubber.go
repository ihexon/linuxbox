//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package vfkit

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

	gvptypes "github.com/containers/gvisor-tap-vsock/pkg/types"
	vfConfig "github.com/crc-org/vfkit/pkg/config"
	"github.com/sirupsen/logrus"
)

type VFkitStubber struct {
	vmconfig.AppleVFkitConfig
}

func (l VFkitStubber) ExtractBootable(userInputPath string, mc *vmconfig.MachineConfig) error {
	destDir := mc.Bootable.Image.GetPath()
	logrus.Infof("Try to decompress %s to %s", userInputPath, destDir)
	if err := decompress.Zstd(userInputPath, mc.Bootable.Image.GetPath()); err != nil {
		errors := fmt.Errorf("could not decompress %s to %s, %w", userInputPath, destDir, err)
		return errors
	}
	return nil
}

func (l VFkitStubber) SetupProviderNetworking(mc *vmconfig.MachineConfig, gvcmd *gvptypes.GvproxyCommand) error {
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

func (l VFkitStubber) MountType() volumes.VolumeMountType {
	return volumes.VirtIOFS
}

func (l VFkitStubber) VMType() defconfig.VMType {
	return defconfig.VFkit
}

func (l VFkitStubber) StartVM(ctx context.Context, mc *vmconfig.MachineConfig) error {
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
	logrus.Infof("vfkit binary path is: %q", vfkitBin)

	cmd, err := vmc.Cmd(vfkitBin)
	if err != nil {
		return fmt.Errorf("failed to create vfkit command: %w", err)
	}

	if err = ignition.GenerateIgnScripts(mc); err != nil {
		return fmt.Errorf("failed to generate ignition scripts: %w", err)
	}

	vfkitCmd := exec.CommandContext(ctx, vfkitBin, cmd.Args[1:]...)

	logrus.Infof("FULL VFKIT CMDLINE: %q", vfkitCmd.Args)
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

// CreateVMConfig generates VM settings aligned with the krunkit structure,
// the structure only used to boot krunkit virtual Machine
func (l VFkitStubber) CreateVMConfig(mc *vmconfig.MachineConfig) error {
	mc.AppleVFkitHypervisor = new(vmconfig.AppleVFkitConfig)
	mc.AppleVFkitHypervisor.Vfkit = vmconfig.Helper{}
	bl := vfConfig.NewEFIBootloader(fmt.Sprintf("%s/efi-bl-%s", mc.Dirs.DataDir.GetPath(), mc.VMName), true)
	mc.AppleVFkitHypervisor.Vfkit.VirtualMachine = vfConfig.NewVirtualMachine(uint(mc.Resources.CPUs), uint64(mc.Resources.Memory), bl)
	randPort, err := port.GetFree(0)
	if err != nil {
		return fmt.Errorf("failed to get random port: %w", err)
	}
	// Endpoint is a string: http://127.0.0.1/[random_port]
	mc.AppleVFkitHypervisor.Vfkit.Endpoint = fmt.Sprintf("%s:%s", define.LocalHostURL, strconv.Itoa(randPort))
	mc.AppleVFkitHypervisor.Vfkit.LogLevel = logrus.InfoLevel
	return nil
}
