//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"bauklotze/pkg/machine/env"

	"github.com/sirupsen/logrus"
)

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
	Provider          string
	HelperBinariesDir Slice
}

func defaultConfig() *Config {
	c := &Config{Machine: defaultMachineConfig()}
	c.Machine.HelperBinariesDir.Set(defaultHelperBinariesDir)
	return c
}

type Config struct {
	Machine MachineConfig `toml:"machine"`
}

const (
	bindirPrefix = "$BINDIR"
)

func findBindir() string {
	execPath, err := os.Executable()
	if err != nil {
		execPath = ""
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		logrus.Warnf("Error resolving symlinks: %v\n", err)
		return ""
	}
	execPath = filepath.Dir(execPath)
	return execPath
}

func (c *Config) FindHelperBinary(name string) (string, error) {
	dirList := c.Machine.HelperBinariesDir.Get()
	if len(dirList) == 0 {
		return "", fmt.Errorf("could not find %q because there are no helper binary directories configured", name)
	}

	for _, path := range dirList {
		if path == bindirPrefix || strings.HasPrefix(path, bindirPrefix+string(filepath.Separator)) {
			// Calculate the path to the executable first time we encounter a $BINDIR prefix.
			bindirPath := findBindir()

			// If there's an error, don't stop the search for the helper binary.
			// findBindir() will have warned once during the first failure.
			if bindirPath == "" {
				return "", fmt.Errorf("failed to find $BINDIR")
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
				err = os.Setenv(env.DYLDLibraryPath, fmt.Sprintf("%s:%s", path, os.Getenv(env.DYLDLibraryPath)))
				if err != nil {
					return "", fmt.Errorf("can not set env DYLD_LIBRARY_PATH with %s", path)
				}
				return lp, nil
			}
		}
	}

	return "", fmt.Errorf("could not find %q in one of %v", name, dirList)
}
