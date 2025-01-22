//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package defconfig

import (
	"bauklotze/pkg/machine/define"
	"runtime"
	"sync"
)

const (
	defaultDiskSize     = 100
	defaultMemory       = 2048
	defaultDataDiskSize = 100
)

var (
	cachedConfig *DefaultConfig
	configOnce   sync.Once
)

type DefaultConfig struct {
	// Number of CPU's a machine is created with.
	CPUs uint64
	// DiskSize is the size of the disk in GB created when init-ing a podman-machine VM
	DiskSize uint64
	// DataDiskSize is the size of the disk in GB created when init-ing virtualMachine mounted to /var
	DataDiskSize uint64
	// Memory in MB a machine is created with.
	Memory uint64
	// Image is the image used when init-ing a podman-machine VM
	Image string
	// Volumes are host directories mounted into the VM by default.
	Volumes Slice
	// Provider is the virtualization provider used to run podman-machine VM
	Provider string
	// Name is the vm name
	Name string
}

func VMConfig() *DefaultConfig {
	configOnce.Do(func() {
		cachedConfig = defaultMachineConfig()
	})
	return cachedConfig
}

func defaultMachineConfig() *DefaultConfig {
	cpus := runtime.NumCPU() / 2 //nolint:mnd
	if cpus == 0 {
		cpus = 1
	}
	return &DefaultConfig{
		CPUs:         uint64(cpus),
		DiskSize:     defaultDiskSize,
		Memory:       defaultMemory,
		DataDiskSize: defaultDataDiskSize,
		Volumes:      NewSlice(getDefaultMachineVolumes()),
		Name:         define.DefaultMachineName,
	}
}
