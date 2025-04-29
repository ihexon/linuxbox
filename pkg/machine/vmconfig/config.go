// SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
// SPDX-License-Identifier: MPL-2.0

package vmconfig

type ResourceConfig struct {
	// CPUs to be assigned to the VM
	CPUs int64 `json:"cpus,omitempty"`
	// DataDisk size in gigabytes assigned to the vm
	DataDiskSizeGB int64 `json:"dataDiskSizeGB,omitempty"`
	// Memory in megabytes assigned to the vm
	MemoryInMB int64 `json:"memory,omitempty"`
}

type VMOpts struct {
	VMName      string
	Workspace   string
	PPID        int64
	CPUs        int64
	MemoryInMiB int64
	Volumes     []string
	BootImage   string
	BootVersion string
	DataVersion string
	ReInit      bool
	ReportURL   string

	// Inject VMType to reduce the number of args
	VMType VMType
}
