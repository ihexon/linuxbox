//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package config

// Destination represents destination for remote service
type Destination struct {
	// URI, required. Example: ssh://root@example.com:22/run/podman/podman.sock
	URI string `json:"URI" toml:"uri"`

	// Identity file with ssh key, optional
	Identity string `json:"Identity,omitempty" toml:"identity,omitempty"`
}

type MachineConfig struct {
	// Number of CPU's a machine is created with.
	CPUs uint64
	// DiskSize is the size of the disk in GB created when init-ing a podman-machine VM
	DiskSize uint64
	// DataDiskSize is the size of the disk in GB created when init-ing virtualMachine mounted to /var
	DataDiskSize uint64
	// Image is the image used when init-ing a podman-machine VM
	Image string
	// Memory in MB a machine is created with.
	Memory uint64
	// User to use for rootless podman when init-ing a podman machine VM
	User string
	// Volumes are host directories mounted into the VM by default.
	Volumes Slice
	// Provider is the virtualization provider used to run podman-machine VM
	Provider string
}

func defaultConfig() *Config {
	return &Config{Machine: defaultMachineConfig()}
}

type Config struct {
	Machine MachineConfig `toml:"machine"`
}
