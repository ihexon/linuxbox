package vmconfigs

import (
	"bauklotze/pkg/lockfile"
	"bauklotze/pkg/machine/lock"
	"bauklotze/pkg/machine/machineDefine"
	"bauklotze/pkg/machine/ports"
	strongunits "bauklotze/pkg/storage"
	"encoding/json"
	"errors"
	"fmt"
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
	VMType() machineDefine.VMType
	Exists(name string) (bool, error)
	GetDisk(userInputPath string, dirs *machineDefine.MachineDirs, mc *MachineConfig) error
	CreateVM(opts machineDefine.CreateVMOpts, mc *MachineConfig) error
	StopVM(mc *MachineConfig, hardStop bool) error
	MountType() VolumeMountType
	RequireExclusiveActive() bool
	State(mc *MachineConfig, bypass bool) (machineDefine.Status, error)
	UpdateSSHPort(mc *MachineConfig, port int) error
	UseProviderNetworkSetup() bool
	StartNetworking(mc *MachineConfig, cmd *gvproxy.GvproxyCommand) error
	PostStartNetworking(mc *MachineConfig, noInfo bool) error
	StartVM(mc *MachineConfig) (func() error, func() error, error)
	MountVolumesToVM(mc *MachineConfig, quiet bool) error
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
func (mc *MachineConfig) DataDir() (*machineDefine.VMFile, error) {
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

	Dirs                   *machineDefine.MachineDirs
	HostUser               HostUser
	Name                   string
	ImagePath              *machineDefine.VMFile
	WSLHypervisor          *WSLConfig `json:",omitempty"`
	ConfigPath             *machineDefine.VMFile
	Resources              machineDefine.ResourceConfig
	imageDescription       machineImage
	Version                uint
	Mounts                 []*Mount
	AppleKrunkitHypervisor *AppleKrunkitConfig `json:",omitempty"`
	GvProxy                gvproxy.GvproxyCommand
	SSH                    SSHConfig
	Starting               bool
	lock                   *lockfile.LockFile
	EvtSockPath            *machineDefine.VMFile `json:",omitempty"`
	TwinPid                int                   `json:",omitempty"`
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
func (mc *MachineConfig) RuntimeDir() (*machineDefine.VMFile, error) {
	if mc.Dirs == nil || mc.Dirs.RuntimeDir == nil {
		return nil, errors.New("no runtime directory set")
	}
	return mc.Dirs.RuntimeDir, nil
}

func (mc *MachineConfig) RemoveRuntimeFiles() ([]string, func() error, error) {
	return nil, nil, nil
}

func NewMachineConfig(opts machineDefine.InitOptions, dirs *machineDefine.MachineDirs, sshIdentityPath string, mtype machineDefine.VMType) (*MachineConfig, error) {
	mc := new(MachineConfig)
	mc.Name = opts.Name
	mc.Dirs = dirs

	// Assign Dirs
	cf, err := machineDefine.NewMachineFile(filepath.Join(dirs.ConfigDir.GetPath(), fmt.Sprintf("%s.json", opts.Name)))
	if err != nil {
		return nil, err
	}
	mc.ConfigPath = cf

	// System Resources
	mrc := machineDefine.ResourceConfig{
		CPUs:     opts.CPUS,
		DiskSize: strongunits.GiB(opts.DiskSize),
		Memory:   strongunits.MiB(opts.Memory),
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

	mc.HostUser = HostUser{UID: getHostUID(), Rootful: opts.Rootful}

	return mc, nil
}

func getHostUID() int {
	return os.Getuid()
}

func loadMachineFromFQPath(path *machineDefine.VMFile) (*MachineConfig, error) {
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

func LoadMachineByName(name string, dirs *machineDefine.MachineDirs) (*MachineConfig, error) {
	fullPath, err := dirs.ConfigDir.AppendToNewVMFile(name + ".json")
	if err != nil {
		return nil, err
	}
	mc, err := loadMachineFromFQPath(fullPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, &machineDefine.ErrVMDoesNotExist{Name: name}
		}
		return nil, err
	}
	mc.Dirs = dirs
	mc.ConfigPath = fullPath

	// If we find an incompatible configuration, we return a hard
	// error because the user wants to deal directly with this
	// machine
	if mc.Version == 0 {
		return mc, &machineDefine.ErrIncompatibleMachineConfig{
			Name: name,
			Path: fullPath.GetPath(),
		}
	}
	return mc, nil
}

func LoadMachinesInDir(dirs *machineDefine.MachineDirs) (map[string]*MachineConfig, error) {
	mcs := make(map[string]*MachineConfig)
	if err := filepath.WalkDir(dirs.ConfigDir.GetPath(), func(path string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(d.Name(), ".json") {
			fullPath, err := dirs.ConfigDir.AppendToNewVMFile(d.Name())
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
				tmpErr := &machineDefine.ErrIncompatibleMachineConfig{
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
