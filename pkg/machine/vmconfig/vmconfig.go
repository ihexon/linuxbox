//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package vmconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	allFlag "bauklotze/pkg/machine/allflag"
	"bauklotze/pkg/machine/defconfig"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/io"
	"bauklotze/pkg/machine/volumes"

	"github.com/containers/common/pkg/strongunits"
	"github.com/containers/storage/pkg/ioutils"
	"github.com/go-playground/validator/v10"

	gvproxy "github.com/containers/gvisor-tap-vsock/pkg/types"
)

type VMProvider interface { //nolint:interfacebloat
	VMType() defconfig.VMType
	ExtractBootable(userInputPath string, mc *MachineConfig) error
	CreateVMConfig(mc *MachineConfig) error
	MountType() volumes.VolumeMountType
	SetupProviderNetworking(mc *MachineConfig, cmd *gvproxy.GvproxyCommand) error
	StartVM(mc *MachineConfig) error
}

func (mc *MachineConfig) PodmanAPISocketHost() *io.VMFile {
	socksDir := mc.Dirs.SocksDir
	s := fmt.Sprintf("%s-podman-api.sock", mc.VMName)
	podmanAPI, _ := socksDir.AppendToNewVMFile(s)
	return podmanAPI
}

// HostUser describes the host user
type HostUser struct {
	UserName string `json:"UserName"`
}

type MachineConfig struct {
	VMProvider VMProvider `json:"-"`
	GvpCmd     *exec.Cmd  `json:"-"`
	VmmCmd     *exec.Cmd  `json:"-"`

	Created  time.Time    `json:"Created"`
	LastUp   time.Time    `json:"LastUp"`
	Dirs     *MachineDirs `json:"Dirs"`
	HostUser HostUser     `json:"HostUser"`
	VMName   string       `json:"Name"`

	Bootable Bootable `json:"Bootable"`
	DataDisk DataDisk `json:"DataDisk"`

	AppleKrunkitHypervisor *AppleKrunkitConfig `json:"AppleKrunkitHypervisor,omitempty"`
	AppleVFkitHypervisor   *AppleVFkitConfig   `json:"AppleVFkitConfig,omitempty"`

	ConfigPath *io.VMFile       `json:"ConfigPath"`
	Resources  ResourceConfig   `json:"Resources"`
	Mounts     []*volumes.Mount `json:"Mounts"`
	GvProxy    GvproxyCommand   `json:"GvProxy"`
	SSH        SSHConfig        `json:"SSH"`
	Starting   bool             `json:"Starting"`
	ReportURL  *io.VMFile       `json:"ReportURL,omitempty"`
}

type Bootable struct {
	Image   *io.VMFile `json:"ImagePath" validate:"required"`
	Version string     `json:"Version"   validate:"required"`
}

type DataDisk struct {
	Image   *io.VMFile `json:"ImagePath" validate:"required"`
	Version string     `json:"Version"   validate:"required"`
}

type GvproxyCommand struct {
	// Print packets on stderr
	Debug bool `json:"Debug,omitempty"`
	// Length of packet
	// Larger packets means less packets to exchange for the same amount of data (and less protocol overhead)
	MTU int `json:"MTU,omitempty"`
	// Values passed in by forward-xxx flags in commandline (forward-xxx:info)
	ForwardInfo map[string][]string `json:"ForwardInfo,omitempty"`
	// List of endpoints the user wants to listen to
	Endpoints []string `json:"Endpoints,omitempty"`
	// Map of different sockets provided by user (socket-type allflag:socket)
	Sockets map[string]string `json:"Sockets,omitempty"`
	// Logfile where gvproxy should redirect logs
	LogFile string `json:"LogFile,omitempty"`
	// File where gvproxy's pid is stored
	PidFile string `json:"PidFile,omitempty"`
	// SSHPort to access the guest VM
	SSHPort int `json:"SSHPort,omitempty"`
	// Podman fordwarding host to guest endpoint, for compatibility
	HostSocks []string `json:"HostSocks"`
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

// NewMachineConfig construct a machine configure but **not* write into disk
func NewMachineConfig(dirs *MachineDirs, sshKey *io.VMFile, mtype defconfig.VMType) (*MachineConfig, error) {
	mc := new(MachineConfig)
	mc.VMName = allFlag.VMName
	mc.Dirs = dirs

	// Assign Dirs
	cf, err := io.NewMachineFile(filepath.Join(dirs.ConfigDir.GetPath(), fmt.Sprintf("%s.json", allFlag.VMName)))
	if err != nil {
		return nil, fmt.Errorf("failed to create machine config file: %w", err)
	}

	mc.ConfigPath = cf
	mc.Resources = ResourceConfig{
		CPUs:           allFlag.CPUS,
		DataDiskSizeGB: define.DataDiskSize, //nolint:mnd
		Memory:         strongunits.MiB(allFlag.Memory),
	}

	mc.SSH = SSHConfig{
		IdentityPath:   sshKey.GetPath(),
		Port:           define.DefaultSSHPort,
		RemoteUsername: define.DefaultUserInVM,
	}
	mc.Created = time.Now()

	u, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get host user information for %w, %s", err, mc.ConfigPath.GetPath())
	}
	mc.HostUser = HostUser{
		UserName: u.Username,
	}

	return mc, nil
}

func LoadMachinesInDir(dirs *MachineDirs) (map[string]*MachineConfig, error) {
	mcs := make(map[string]*MachineConfig)
	err := filepath.WalkDir(dirs.ConfigDir.GetPath(), func(path string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(d.Name(), ".json") {
			fullPath, err := dirs.ConfigDir.AppendToNewVMFile(d.Name())
			if err != nil {
				return fmt.Errorf("failed to create full path: %w", err)
			}
			mc, err := loadMachineFromFQPath(fullPath)
			if err != nil {
				return fmt.Errorf("failed to load machine config: %w", err)
			}

			mc.ConfigPath = fullPath
			mc.Dirs = dirs
			mcs[mc.VMName] = mc
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk machine config dir: %w", err)
	}

	return mcs, nil
}

// LoadMachineByName returns a machine config based on the vm name and provider
func LoadMachineByName(name string, dirs *MachineDirs) (*MachineConfig, error) {
	fullPath, err := dirs.ConfigDir.AppendToNewVMFile(name + ".json")
	if err != nil {
		return nil, fmt.Errorf("error in LoadMachineByName, AppendToNewVMFile failed: %w", err)
	}

	mc, err := loadMachineFromFQPath(fullPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("VM does not exist")
		}
		return nil, err
	}
	mc.Dirs = dirs
	mc.ConfigPath = fullPath

	return mc, nil
}

func loadMachineFromFQPath(f *io.VMFile) (*MachineConfig, error) {
	mc := new(MachineConfig)
	b, err := f.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read machine config: %w", err)
	}

	if err = json.Unmarshal(b, mc); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	validate := validator.New()
	if err = validate.Struct(mc); err != nil {
		return nil, fmt.Errorf("invalid machine config: %w", err)
	}

	return mc, nil
}

func (mc *MachineConfig) GVProxyNetworkBackendSocks() (*io.VMFile, error) {
	socksDir, err := mc.SocksDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace tmp dir: %w", err)
	}
	return socksDir.AppendToNewVMFile(fmt.Sprintf("%s-gvproxy.sock", mc.VMName)) //nolint:wrapcheck
}

// SocksDir is simple helper function to obtain the workspace tmp dir
func (mc *MachineConfig) SocksDir() (*io.VMFile, error) {
	if mc.Dirs == nil || mc.Dirs.SocksDir.GetPath() == "" {
		return nil, errors.New("no workspace socks directory set")
	}
	return mc.Dirs.SocksDir, nil
}

// write is a non-locking way to write the machine configuration file to disk
func (mc *MachineConfig) Write() error {
	if mc.ConfigPath == nil {
		return fmt.Errorf("no configuration file associated with vm %q", mc.VMName)
	}
	b, err := json.Marshal(mc)
	if err != nil {
		return fmt.Errorf("failed to marshal machine config: %w", err)
	}
	return ioutils.AtomicWriteFile(mc.ConfigPath.GetPath(), b, define.DefaultFilePerm) //nolint:wrapcheck
}
