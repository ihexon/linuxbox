//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package vfkit

import (
	"context"

	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/vmconfig"
	"bauklotze/pkg/machine/volumes"
)

type VFkitStubber struct {
	VMState *vmconfig.VMState
}

func NewProvider() *VFkitStubber {
	return &VFkitStubber{
		VMState: &vmconfig.VMState{
			SSHReady:    false,
			PodmanReady: false,
		},
	}
}

func (l VFkitStubber) MountType() volumes.VolumeMountType {
	return volumes.VirtIOFS
}

func (l VFkitStubber) InitializeVM(opts vmconfig.VMOpts) (*vmconfig.MachineConfig, error) {
	return machine.InitializeVM(opts) //nolint:wrapcheck
}

func (l VFkitStubber) VMType() vmconfig.VMType {
	return vmconfig.VFkit
}

func (l VFkitStubber) StartNetworkProvider(ctx context.Context, mc *vmconfig.MachineConfig) error {
	return nil
}

func (l VFkitStubber) StartVMProvider(ctx context.Context, mc *vmconfig.MachineConfig) error {
	return nil
}

func (l VFkitStubber) StartSSHAuthService(ctx context.Context, mc *vmconfig.MachineConfig) error {
	return nil
}

func (l VFkitStubber) StartTimeSyncService(ctx context.Context, mc *vmconfig.MachineConfig) error {
	return nil
}

func (l VFkitStubber) GetVMState() *vmconfig.VMState {
	return l.VMState
}
