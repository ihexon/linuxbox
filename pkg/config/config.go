package config

import (
	"github.com/spf13/pflag"
)

// This is a higher-level configuration structure that includes more general settings for the virtual machine.
// It's used for overall machine management and configuration.
type MachineConfig struct {
	// Number of CPU's a machine is created with.
	CPUs uint64 `toml:"cpus,omitempty,omitzero"`
	// DiskSize is the size of the disk in GB created when init-ing a podman-machine VM
	DiskSize uint64 `toml:"disk_size,omitempty,omitzero"`
	// Image is the image used when init-ing a podman-machine VM
	Image string `toml:"image,omitempty"`
	// Memory in MB a machine is created with.
	Memory uint64 `toml:"memory,omitempty,omitzero"`
	// User to use for rootless podman when init-ing a podman machine VM
	User string `toml:"user,omitempty"`
	// Volumes are host directories mounted into the VM by default.
	Volumes Slice `toml:"volumes,omitempty"`
	// Provider is the virtualization provider used to run podman-machine VM
	Provider string `toml:"provider,omitempty"`
}

type Config struct {
	Machine MachineConfig `toml:"machine"`
}

type PodmanConfig struct {
	*pflag.FlagSet
	ContainersConf           *Config
	ContainersConfDefaultsRO *Config
}
