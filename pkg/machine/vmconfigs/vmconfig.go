//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package vmconfigs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/lock"
	"bauklotze/pkg/machine/ports"

	"github.com/containers/common/pkg/strongunits"
	gvproxy "github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/containers/storage/pkg/lockfile"
	"github.com/sirupsen/logrus"
)

type VMProvider interface { //nolint:interfacebloat
	VMType() define.VMType
	GetDisk(userInputPath string, dirs *define.MachineDirs, imagePath *define.VMFile, vmType define.VMType, name string) error
	CreateVM(opts define.CreateVMOpts, mc *MachineConfig) error
	MountType() VolumeMountType
	State(mc *MachineConfig) (define.Status, error)
	StartNetworking(mc *MachineConfig, cmd *gvproxy.GvproxyCommand) error
	StartVM(mc *MachineConfig) (*exec.Cmd, func() error, error)
}

type Mount struct {
	OriginalInput string  `json:"OriginalInput"`
	ReadOnly      bool    `json:"ReadOnly"`
	Source        string  `json:"Source"`
	Tag           string  `json:"Tag"`
	Target        string  `json:"Target"`
	Type          string  `json:"Type"`
	VSockNumber   *uint64 `json:"VSockNumber"`
}

// HostUser describes the host user
type HostUser struct {
	// Whether this machine should run in a rootful or rootless manner
	Rootful bool `json:"Rootful"`
	// UID is the numerical id of the user that called machine
	UID int `json:"UID"`
	// Whether one of these fields has changed and actions should be taken
	Modified bool `json:"HostUserModified"`
}

type MachineConfig struct {
	Created time.Time `json:"Created"`
	LastUp  time.Time `json:"LastUp"`

	Dirs     *define.MachineDirs `json:"Dirs"`
	HostUser HostUser            `json:"HostUser"`
	Name     string              `json:"Name"`
	// TODO Using Image struct and BootableDiskVersion struct
	ImagePath   *define.VMFile `json:"ImagePath"`   // mc.ImagePath is the bootable copied from user provided image --boot <bootable.img.xz>
	DataDisk    *define.VMFile `json:"DataDisk"`    // External Disk file
	OverlayDisk *define.VMFile `json:"OverlayDisk"` // Overlay Disk file

	BootableDiskVersion string `json:"BootableDiskVersion,omitempty"` // Bootable Image for now
	DataDiskVersion     string `json:"DataDiskVersion,omitempty"`     // External Disk for now

	AppleKrunkitHypervisor *AppleKrunkitConfig `json:"AppleKrunkitHypervisor,omitempty"`
	AppleVFkitHypervisor   *AppleVFkitConfig   `json:"AppleVFkitConfig,omitempty"`

	ConfigPath *define.VMFile        `json:"ConfigPath"`
	Resources  define.ResourceConfig `json:"Resources"`
	Version    uint                  `json:"Version"`
	Mounts     []*Mount              `json:"Mounts"`
	GvProxy    GvproxyCommand        `json:"GvProxy"`
	SSH        SSHConfig             `json:"SSH"`
	Starting   bool                  `json:"Starting"`
	lock       *lockfile.LockFile
	// Oomol Studio
	ReportURL *define.VMFile `json:",omitempty"`
}

type GvproxyCommand struct {
	GvProxy     gvproxy.GvproxyCommand `json:"GvProxy"`
	HostSocks   []string               `json:"HostSocks"`
	RemoteSocks string                 `json:"RemoteSocks"`
}

// SSHConfig contains remote access information for SSH
type SSHConfig struct {
	// IdentityPath is the fq path to the ssh priv key
	IdentityPath string `json:"IdentityPath"`
	// SSH port for user networking
	Port int `json:"Port"`
	// RemoteUsername of the vm user
	RemoteUsername string `json:"RemoteUsername"`
}

// RuntimeDir is simple helper function to obtain the runtime dir
func (mc *MachineConfig) RuntimeDir() (*define.VMFile, error) {
	if mc.Dirs == nil || mc.Dirs.RuntimeDir == nil {
		return nil, errors.New("no runtime directory set")
	}
	return mc.Dirs.RuntimeDir, nil
}

func NewMachineConfig(opts define.InitOptions, dirs *define.MachineDirs, sshIdentityPath string, mtype define.VMType) (*MachineConfig, error) {
	mc := new(MachineConfig)
	mc.Name = opts.Name
	mc.Dirs = dirs

	// Assign Dirs
	cf, err := define.NewMachineFile(filepath.Join(dirs.ConfigDir.GetPath(), fmt.Sprintf("%s.json", opts.Name)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create machine config file: %w", err)
	}
	mc.ConfigPath = cf

	// System Resources
	mrc := define.ResourceConfig{
		CPUs: opts.CPUS,
		// DiskSize: strongunits.GiB(opts.DiskSize),
		Memory: strongunits.MiB(opts.Memory),
	}
	mc.Resources = mrc

	var sshPort int
	listener, tempErr := net.Listen("tcp", "127.0.0.1:61234")
	if tempErr != nil {
		logrus.Infof("Gvproxy SSH port 61234 port can not be used , try to get a random port...")
		if sshPort, err = ports.AllocateMachinePort(); err != nil {
			return nil, fmt.Errorf("failed to allocate machine port: %w", err)
		}
	} else {
		_, portString, _ := net.SplitHostPort(listener.Addr().String())
		sshPort, _ = strconv.Atoi(portString)
		listener.Close()
	}

	sshConfig := SSHConfig{
		IdentityPath:   sshIdentityPath,
		Port:           sshPort,
		RemoteUsername: opts.Username, // always be root
	}
	mc.SSH = sshConfig
	mc.Created = time.Now()

	mc.HostUser = HostUser{
		UID:     getHostUID(),
		Rootful: true, // Default root
	}

	return mc, nil
}

func getHostUID() int {
	return os.Getuid()
}

func LoadMachineByName(name string, dirs *define.MachineDirs) (*MachineConfig, error) {
	fullPath, err := dirs.ConfigDir.AppendToNewVMFile(name+".json", nil)
	logrus.Infof("Try load MachineConfigure %s from %s", name, fullPath.GetPath())
	if err != nil {
		return nil, fmt.Errorf("failed to create full path: %w", err)
	}
	mc, err := loadMachineFromFQPath(fullPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, &define.VMDoesNotExistError{Name: name}
		}
		return nil, fmt.Errorf("failed to load machine config: %w", err)
	}
	mc.Dirs = dirs
	mc.ConfigPath = fullPath

	// If we find an incompatible configuration, we return a hard
	// error because the user wants to deal directly with this
	// machine
	if mc.Version == 0 {
		return mc, &define.IncompatibleMachineConfigError{
			Name: name,
			Path: fullPath.GetPath(),
		}
	}
	return mc, nil
}

func LoadMachinesInDir(dirs *define.MachineDirs) (map[string]*MachineConfig, error) {
	mcs := make(map[string]*MachineConfig)
	err := filepath.WalkDir(dirs.ConfigDir.GetPath(), func(path string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(d.Name(), ".json") {
			fullPath, err := dirs.ConfigDir.AppendToNewVMFile(d.Name(), nil)
			if err != nil {
				return fmt.Errorf("failed to create full path: %w", err)
			}
			mc, err := loadMachineFromFQPath(fullPath)
			if err != nil {
				return fmt.Errorf("failed to load machine config: %w", err)
			}
			// if we find an incompatible machine configuration file, we emit and error
			if mc.Version == 0 {
				tmpErr := &define.IncompatibleMachineConfigError{
					Name: mc.Name,
					Path: fullPath.GetPath(),
				}
				logrus.Error(tmpErr)
				return nil
			}
			mc.ConfigPath = fullPath
			mc.Dirs = dirs
			mcs[mc.Name] = mc
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk machine config dir: %w", err)
	}

	return mcs, nil
}

func loadMachineFromFQPath(f *define.VMFile) (*MachineConfig, error) {
	mc := new(MachineConfig)
	b, err := f.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read machine config: %w", err)
	}

	if err = json.Unmarshal(b, mc); err != nil {
		return nil, fmt.Errorf("unable to load machine config file: %w", err)
	}
	lock, err := lock.GetMachineLock(mc.Name, filepath.Dir(f.GetPath()))
	if err != nil {
		return nil, fmt.Errorf("failed to get machine lock: %w", err)
	}
	mc.lock = lock
	return mc, nil
}
