//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package krunkit

import (
	"fmt"
	"os/exec"

	"bauklotze/pkg/machine/apple/hvhelper"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/diskpull"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/port"

	gvproxy "github.com/containers/gvisor-tap-vsock/pkg/types"
	vfConfig "github.com/crc-org/vfkit/pkg/config"
	"github.com/sirupsen/logrus"
)

type LibKrunStubber struct {
	vmconfigs.AppleKrunkitConfig
}

func (l LibKrunStubber) State(mc *vmconfigs.MachineConfig) (define.Status, error) {
	return mc.AppleKrunkitHypervisor.Krunkit.State() //nolint:wrapcheck
}

func (l LibKrunStubber) GetDisk(userInputPath string, dirs *define.MachineDirs, imagePath *define.VMFile, vmType define.VMType, name string) error {
	// mc.ImagePath is the bootable copied from user provided image --boot <bootable.img.xz>
	// userInputPath is the bootable image user provided
	// Extract  userInputPath --> imagePath
	return diskpull.GetDisk(userInputPath, imagePath) //nolint:wrapcheck
}

func (l LibKrunStubber) StartNetworking(mc *vmconfigs.MachineConfig, cmd *gvproxy.GvproxyCommand) error {
	return StartGenericNetworking(mc, cmd)
}

func (l LibKrunStubber) MountType() vmconfigs.VolumeMountType {
	return vmconfigs.VirtIOFS
}

const (
	krunkitBinary = "krunkit"
	localhostURI  = "http://127.0.0.1"
)

func (l LibKrunStubber) CreateVM(opts define.CreateVMOpts, mc *vmconfigs.MachineConfig) error {
	mc.AppleKrunkitHypervisor = new(vmconfigs.AppleKrunkitConfig)
	mc.AppleKrunkitHypervisor.Krunkit = hvhelper.Helper{}

	bl := vfConfig.NewEFIBootloader(fmt.Sprintf("%s/efi-bl-%s", opts.Dirs.DataDir.GetPath(), opts.Name), true)
	mc.AppleKrunkitHypervisor.Krunkit.VirtualMachine = vfConfig.NewVirtualMachine(uint(mc.Resources.CPUs), uint64(mc.Resources.Memory), bl)

	randPort, err := port.GetFree(0)
	if err != nil {
		return fmt.Errorf("failed to get random port: %w", err)
	}

	mc.AppleKrunkitHypervisor.Krunkit.Endpoint = fmt.Sprintf("%s:%d", localhostURI, randPort)
	mc.AppleKrunkitHypervisor.Krunkit.LogLevel = logrus.InfoLevel

	return nil
}

func (l LibKrunStubber) VMType() define.VMType {
	return define.LibKrun
}

func (l LibKrunStubber) StartVM(mc *vmconfigs.MachineConfig) (*exec.Cmd, func() error, error) {
	bl := mc.AppleKrunkitHypervisor.Krunkit.VirtualMachine.Bootloader
	if bl == nil {
		return nil, nil, fmt.Errorf("unable to determine boot loader for this machine")
	}
	return StartGenericAppleVM(mc, krunkitBinary, bl, mc.AppleKrunkitHypervisor.Krunkit.Endpoint)
}
