//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package vfkit

import (
	"fmt"
	"strconv"

	"bauklotze/pkg/decompress"
	"bauklotze/pkg/machine/defconfig"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfig"
	"bauklotze/pkg/machine/volumes"
	"bauklotze/pkg/port"

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

func (l VFkitStubber) StartVM(mc *vmconfig.MachineConfig) error {
	return startVFKit(mc)
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
