//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package krunkit

import (
	"bauklotze/pkg/decompress"
	"bauklotze/pkg/machine/defconfig"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfig"
	"bauklotze/pkg/machine/volumes"
	"bauklotze/pkg/port"
	"fmt"
	"strconv"

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

func (l LibKrunStubber) StartVM(mc *vmconfig.MachineConfig) error {
	return startKrunkit(mc)
}
