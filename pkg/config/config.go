package config

import (
	"github.com/spf13/pflag"
)

type FarmConfig struct {
	// Default is the default farm to be used when farming out builds
	Default string `json:",omitempty" toml:"default,omitempty"`
	// List is a map of farms created where key=farm-name and value=list of connections
	List map[string][]string `json:",omitempty" toml:"list,omitempty"`
}

type ConnectionConfig struct {
	Default     string                 `json:",omitempty"`
	Connections map[string]Destination `json:",omitempty"`
}

// Destination represents destination for remote service
type Destination struct {
	// URI, required. Example: ssh://root@example.com:22/run/podman/podman.sock
	URI string `toml:"uri"`

	// Identity file with ssh key, optional
	Identity string `json:",omitempty" toml:"identity,omitempty"`

	// isMachine describes if the remote destination is a machine.
	IsMachine bool `json:",omitempty" toml:"is_machine,omitempty"`
}

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
	// HelperPath is the
	HelperBinariesDir Slice `toml:"helper_path,omitempty"`
}

type Config struct {
	Machine MachineConfig `toml:"machine"`
}

type PodmanConfig struct {
	*pflag.FlagSet
	ContainersConf           *Config
	ContainersConfDefaultsRO *Config
}
