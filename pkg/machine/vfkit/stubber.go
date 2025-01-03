//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package vfkit

import (
	"fmt"
	"os/exec"
	"strconv"

	"bauklotze/pkg/machine/apple/hvhelper"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/diskpull"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/port"

	gvproxy "github.com/containers/gvisor-tap-vsock/pkg/types"
	vfConfig "github.com/crc-org/vfkit/pkg/config"
	"github.com/sirupsen/logrus"
)

type VFkitStubber struct {
	vmconfigs.AppleVFkitConfig
}

func (l VFkitStubber) State(mc *vmconfigs.MachineConfig) (define.Status, error) {
	return mc.AppleVFkitHypervisor.Vfkit.State() //nolint:wrapcheck
}

func (l VFkitStubber) GetDisk(userInputPath string, dirs *define.MachineDirs, imagePath *define.VMFile, vmType define.VMType, name string) error {
	// mc.ImagePath is the bootable copied from user provided image --boot <bootable.img.xz>
	// userInputPath is the bootable image user provided
	// Extract  userInputPath --> imagePath
	return diskpull.GetDisk(userInputPath, imagePath) //nolint:wrapcheck
}

func (l VFkitStubber) StartNetworking(mc *vmconfigs.MachineConfig, cmd *gvproxy.GvproxyCommand) error {
	return StartGenericNetworking(mc, cmd)
}

func (l VFkitStubber) MountType() vmconfigs.VolumeMountType {
	return vmconfigs.VirtIOFS
}

const (
	krunkitBinary = "vfkit"
	localhostURI  = "http://127.0.0.1"
)

func (l VFkitStubber) CreateVM(opts define.CreateVMOpts, mc *vmconfigs.MachineConfig) error {
	mc.AppleVFkitHypervisor = new(vmconfigs.AppleVFkitConfig)
	mc.AppleVFkitHypervisor.Vfkit = hvhelper.Helper{}
	bl := vfConfig.NewEFIBootloader(fmt.Sprintf("%s/efi-bl-%s", opts.Dirs.DataDir.GetPath(), opts.Name), true)
	mc.AppleVFkitHypervisor.Vfkit.VirtualMachine = vfConfig.NewVirtualMachine(uint(mc.Resources.CPUs), uint64(mc.Resources.Memory), bl)
	randPort, err := port.GetFree(0)
	if err != nil {
		return fmt.Errorf("failed to get random port: %w", err)
	}
	// Endpoint is a string: http://127.0.0.1/[random_port]
	mc.AppleVFkitHypervisor.Vfkit.Endpoint = localhostURI + ":" + strconv.Itoa(randPort)
	mc.AppleVFkitHypervisor.Vfkit.LogLevel = logrus.InfoLevel

	return nil
}

func (l VFkitStubber) VMType() define.VMType {
	return define.VFkit
}

func (l VFkitStubber) StartVM(mc *vmconfigs.MachineConfig) (*exec.Cmd, func() error, error) {
	bl := mc.AppleVFkitHypervisor.Vfkit.VirtualMachine.Bootloader
	if bl == nil {
		return nil, nil, fmt.Errorf("unable to determine boot loader for this machine")
	}
	return StartGenericAppleVM(mc, krunkitBinary, bl, mc.AppleVFkitHypervisor.Vfkit.Endpoint)
}
