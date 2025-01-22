// SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
// SPDX-License-Identifier: MPL-2.0
package vmconfig

import (
	"bauklotze/pkg/machine/io"

	"github.com/containers/common/pkg/strongunits"
)

type CreateVMOpts struct {
	Name                  string       `json:"Name"`
	Dirs                  *MachineDirs `json:"Dirs"`
	UserProvidedImageFile string       `json:"UserImageFile"` // Only used in wsl2
}

type ResourceConfig struct {
	// CPUs to be assigned to the VM
	CPUs uint64 `json:"CPUs,omitempty"`
	// DataDisk size in gigabytes assigned to the vm
	DataDiskSizeGB strongunits.GiB `json:"DataDiskSizeGB,omitempty"`
	// Memory in megabytes assigned to the vm
	Memory strongunits.MiB `json:"Memory,omitempty"`
	// Usbs
}

type MachineDirs struct {
	ConfigDir       *io.VMFile       `json:"ConfigDir"`
	DataDir         *io.VMFile       `json:"DataDir"`
	TmpDir          *io.VMFile       `json:"RuntimeDir"`
	LogsDir         *io.VMFile       `json:"LogsDir"`
	Hypervisor      *Hypervisor      `json:"Hypervisor"`
	NetworkProvider *NetworkProvider `json:"NetworkProvider"`
}

type Hypervisor struct {
	LibsDir *io.VMFile `json:"LibsDir"`
	Bin     *io.VMFile `json:"Bin"`
}

type NetworkProvider struct {
	LibsDir *io.VMFile `json:"LibsDir"`
	Bin     *io.VMFile `json:"Bin"`
}
