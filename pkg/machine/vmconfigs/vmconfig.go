package vmconfigs

import (
	"bauklotze/pkg/lockfile"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/lock"
	"bauklotze/pkg/machine/ports"
	"encoding/json"
	"errors"
	"fmt"
	strongunits "github.com/containers/common/pkg/strongunits"
	gvproxy "github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func NormalizeMachineArch(arch string) (string, error) {
	switch arch {
	case "arm64", "aarch64":
		return "aarch64", nil
	case "x86_64", "amd64":
		return "x86_64", nil
	}
	return "", fmt.Errorf("unsupported platform %s", arch)
}

type VMProvider interface { //nolint:interfacebloat
	VMType() define.VMType
	Exists(name string) (bool, error)
	GetDisk(userInputPath string, dirs *define.MachineDirs, imagePath *define.VMFile, vmType define.VMType, name string) error
	CreateVM(opts define.CreateVMOpts, mc *MachineConfig) error
	StopVM(mc *MachineConfig, hardStop bool) error
	MountType() VolumeMountType
	RequireExclusiveActive() bool
	State(mc *MachineConfig) (define.Status, error)
	UpdateSSHPort(mc *MachineConfig, port int) error
	StartNetworking(mc *MachineConfig, cmd *gvproxy.GvproxyCommand) error
	PostStartNetworking(mc *MachineConfig, noInfo bool) error
	StartVM(mc *MachineConfig) (func() error, func() error, error)
	MountVolumesToVM(mc *MachineConfig, quiet bool) error
	SetProviderAttrs(mc *MachineConfig, opts define.SetOptions) error
}

type machineImage interface { //nolint:unused
	download() error
	path() string
}

type Mount struct {
	OriginalInput string
	ReadOnly      bool
	Source        string
	Tag           string
	Target        string
	Type          string
	VSockNumber   *uint64
}

// HostUser describes the host user
type HostUser struct {
	// Whether this machine should run in a rootful or rootless manner
	Rootful bool
	// UID is the numerical id of the user that called machine
	UID int
	// Whether one of these fields has changed and actions should be taken
	Modified bool `json:"HostUserModified"`
}

// DataDir is a simple helper function to obtain the machine data dir
func (mc *MachineConfig) DataDir() (*define.VMFile, error) {
	if mc.Dirs == nil || mc.Dirs.DataDir == nil {
		return nil, errors.New("no data directory set")
	}
	return mc.Dirs.DataDir, nil
}

func (mc *MachineConfig) IsFirstBoot() (bool, error) {
	never, err := time.Parse(time.RFC3339, "0001-01-01T00:00:00Z")
	if err != nil {
		return false, err
	}
	return mc.LastUp == never, nil
}

type MachineConfig struct {
	Created time.Time
	LastUp  time.Time

	Dirs         *define.MachineDirs
	HostUser     HostUser
	Name         string
	ImagePath    *define.VMFile // Bootable Image file
	ImageVersion string         `json:",omitempty"` // Bootable Image for now

	ExternalDisk        *define.VMFile // External Disk file
	ExternalDiskVersion string         `json:",omitempty"` // External Disk for now

	AppleKrunkitHypervisor *AppleKrunkitConfig `json:",omitempty"`
	WSLHypervisor          *WSLConfig          `json:",omitempty"`

	ConfigPath *define.VMFile
	Resources  define.ResourceConfig
	Version    uint
	Mounts     []*Mount
	GvProxy    GvproxyCommand
	SSH        SSHConfig
	Starting   bool
	lock       *lockfile.LockFile
	// Oomol Studio
	EvtSockPath *define.VMFile `json:",omitempty"`
}

type GvproxyCommand struct {
	GvProxy     gvproxy.GvproxyCommand
	HostSocks   []string
	RemoteSocks string
}

// SSHConfig contains remote access information for SSH
type SSHConfig struct {
	// IdentityPath is the fq path to the ssh priv key
	IdentityPath string
	// SSH port for user networking
	Port int
	// RemoteUsername of the vm user
	RemoteUsername string
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
		return nil, err
	}
	mc.ConfigPath = cf

	// System Resources
	mrc := define.ResourceConfig{
		CPUs: opts.CPUS,
		//DiskSize: strongunits.GiB(opts.DiskSize),
		Memory: strongunits.MiB(opts.Memory),
	}
	mc.Resources = mrc

	sshPort, err := ports.AllocateMachinePort()
	if err != nil {
		return nil, err
	}

	sshConfig := SSHConfig{
		IdentityPath:   sshIdentityPath,
		Port:           sshPort,
		RemoteUsername: opts.Username,
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
	if err != nil {
		return nil, err
	}
	mc, err := loadMachineFromFQPath(fullPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, &define.ErrVMDoesNotExist{Name: name}
		}
		return nil, err
	}
	mc.Dirs = dirs
	mc.ConfigPath = fullPath

	// If we find an incompatible configuration, we return a hard
	// error because the user wants to deal directly with this
	// machine
	if mc.Version == 0 {
		return mc, &define.ErrIncompatibleMachineConfig{
			Name: name,
			Path: fullPath.GetPath(),
		}
	}
	return mc, nil
}

func LoadMachinesInDir(dirs *define.MachineDirs) (map[string]*MachineConfig, error) {
	mcs := make(map[string]*MachineConfig)
	if err := filepath.WalkDir(dirs.ConfigDir.GetPath(), func(path string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(d.Name(), ".json") {
			fullPath, err := dirs.ConfigDir.AppendToNewVMFile(d.Name(), nil)
			if err != nil {
				return err
			}
			mc, err := loadMachineFromFQPath(fullPath)
			if err != nil {
				return err
			}
			// if we find an incompatible machine configuration file, we emit and error
			//
			if mc.Version == 0 {
				tmpErr := &define.ErrIncompatibleMachineConfig{
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
	}); err != nil {
		return nil, err
	}
	return mcs, nil
}

func loadMachineFromFQPath(path *define.VMFile) (*MachineConfig, error) {
	mc := new(MachineConfig)
	b, err := path.Read()
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(b, mc); err != nil {
		return nil, fmt.Errorf("unable to load machine config file: %q", err)
	}
	lock, err := lock.GetMachineLock(mc.Name, filepath.Dir(path.GetPath()))
	mc.lock = lock
	return mc, err
}
