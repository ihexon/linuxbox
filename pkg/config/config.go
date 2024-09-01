package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

const (
	bindirPrefix = "$BINDIR"
)

var (
	bindirFailed = false
	bindirCached = ""
)

func findBindir() string {
	if bindirCached != "" || bindirFailed {
		return bindirCached
	}
	execPath, err := os.Executable()
	if err == nil {
		// Resolve symbolic links to find the actual binary file path.
		execPath, err = filepath.EvalSymlinks(execPath)
	}
	if err != nil {
		// If failed to find executable (unlikely to happen), warn about it.
		// The bindirFailed flag will track this, so we only warn once.
		logrus.Warnf("Failed to find $BINDIR: %v", err)
		bindirFailed = true
		return ""
	}
	bindirCached = filepath.Dir(execPath)
	return bindirCached
}

func (c *Config) FindHelperBinary(name string, searchPATH bool) (string, error) {
	dirList := c.Machine.HelperBinariesDir.Get()
	bindirPath := ""
	bindirSearched := false

	// If set, search this directory first. This is used in testing.
	if dir, found := os.LookupEnv("CONTAINERS_HELPER_BINARY_DIR"); found {
		dirList = append([]string{dir}, dirList...)
	}

	for _, path := range dirList {
		if path == bindirPrefix || strings.HasPrefix(path, bindirPrefix+string(filepath.Separator)) {
			// Calculate the path to the executable first time we encounter a $BINDIR prefix.
			if !bindirSearched {
				bindirSearched = true
				bindirPath = findBindir()
			}
			// If there's an error, don't stop the search for the helper binary.
			// findBindir() will have warned once during the first failure.
			if bindirPath == "" {
				continue
			}
			// Replace the $BINDIR prefix with the path to the directory of the current binary.
			if path == bindirPrefix {
				path = bindirPath
			} else {
				path = filepath.Join(bindirPath, strings.TrimPrefix(path, bindirPrefix+string(filepath.Separator)))
			}
		}
		// Absolute path will force exec.LookPath to check for binary existence instead of lookup everywhere in PATH
		if abspath, err := filepath.Abs(filepath.Join(path, name)); err == nil {
			// exec.LookPath from absolute path on Unix is equal to os.Stat + IsNotDir + check for executable bits in FileMode
			// exec.LookPath from absolute path on Windows is equal to os.Stat + IsNotDir for `file.ext` or loops through extensions from PATHEXT for `file`
			if lp, err := exec.LookPath(abspath); err == nil {
				return lp, nil
			}
		}
	}
	if searchPATH {
		return exec.LookPath(name)
	}
	configHint := "To resolve this error, set the helper_binaries_dir key in the `[engine]` section of containers.conf to the directory containing your helper binaries."
	if len(dirList) == 0 {
		return "", fmt.Errorf("could not find %q because there are no helper binary directories configured.  %s", name, configHint)
	}
	return "", fmt.Errorf("could not find %q in one of %v.  %s", name, dirList, configHint)
}
