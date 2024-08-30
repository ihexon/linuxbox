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
	// HelperPath is the
	HelperBinariesDir Slice `toml:"helper_path,omitempty"`
}

type Config struct {
	Machine MachineConfig `toml:"machine"`
}

//func (c *Config) FindHelperBinary(name string, searchPATH bool) (string, error) {
//	dirList := c.Machine.HelperBinariesDir.Get()
//	bindirPath := ""
//	bindirSearched := false
//	// If set, search this directory first. This is used in testing.
//	if dir, found := os.LookupEnv("CONTAINERS_HELPER_BINARY_DIR"); found {
//		dirList = append([]string{dir}, dirList...)
//	}
//
//}

type PodmanConfig struct {
	*pflag.FlagSet
	ContainersConf           *Config
	ContainersConfDefaultsRO *Config
}
