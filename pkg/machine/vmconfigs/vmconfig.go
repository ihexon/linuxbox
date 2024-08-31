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
	"io/fs"
	"path/filepath"
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

type MachineConfig struct {
	Created                time.Time
	Dirs                   *machineDefine.MachineDirs
	Name                   string
	ImagePath              *machineDefine.VMFile
	WSLHypervisor          *WSLConfig `json:",omitempty"`
	ConfigPath             *machineDefine.VMFile
	Resources              machineDefine.ResourceConfig
	imageDescription       machineImage
	Version                uint
	Mounts                 []*Mount
	AppleHypervisor        *AppleVfkitConfig   `json:",omitempty"`
	AppleKrunkitHypervisor *AppleKrunkitConfig `json:",omitempty"`
	GvProxy                gvproxy.GvproxyCommand
	SSH                    SSHConfig
	Starting               bool
	lock                   *lockfile.LockFile
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
	return mc, nil
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

// The LoadMachinesInDir function loads all machine configurations from a specified directory.
// It walks through the directory, identifies JSON configuration files, and loads each machine configuration using loadMachineFromFQPath.
// It also handles incompatible machine configurations by logging an error.
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
