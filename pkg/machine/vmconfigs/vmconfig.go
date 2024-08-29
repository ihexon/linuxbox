package vmconfigs

import (
	"bauklotze/pkg/machine/define"
	strongunits "bauklotze/pkg/storage"
	"encoding/json"
	"errors"
	"fmt"
	gvproxy "github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/containers/storage/pkg/lockfile"
	"github.com/sirupsen/logrus"
	"io/fs"
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
	GetDisk(userInputPath string, dirs *define.MachineDirs, mc *MachineConfig) error
	CreateVM(opts define.CreateVMOpts, mc *MachineConfig) error
	StopVM(mc *MachineConfig, hardStop bool) error
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
	Dirs                   *define.MachineDirs
	Name                   string
	ImagePath              *define.VMFile // 实际上是 rootfs 的路径
	WSLHypervisor          *WSLConfig     `json:",omitempty"`
	ConfigPath             *define.VMFile
	Resources              define.ResourceConfig
	imageDescription       machineImage
	Version                uint
	Lock                   *lockfile.LockFile
	Mounts                 []*Mount
	AppleHypervisor        *AppleVfkitConfig   `json:",omitempty"`
	AppleKrunkitHypervisor *AppleKrunkitConfig `json:",omitempty"`
	GvProxy                gvproxy.GvproxyCommand
	Starting               bool
}

// RuntimeDir is simple helper function to obtain the runtime dir
func (mc *MachineConfig) RuntimeDir() (*define.VMFile, error) {
	if mc.Dirs == nil || mc.Dirs.RuntimeDir == nil {
		return nil, errors.New("no runtime directory set")
	}
	return mc.Dirs.RuntimeDir, nil
}

func (mc *MachineConfig) RemoveRuntimeFiles() ([]string, func() error, error) {
	return nil, nil, nil
}

func NewMachineConfig(opts define.InitOptions, dirs *define.MachineDirs, vmtype define.VMType) (*MachineConfig, error) {
	mc := new(MachineConfig)
	mc.Name = opts.Name
	mc.Dirs = dirs

	// Assign Dirs
	cf, err := define.NewMachineFile(filepath.Join(dirs.ConfigDir.GetPath(), fmt.Sprintf("%s.json", opts.Name)))
	if err != nil {
		return nil, err
	}
	mc.ConfigPath = cf

	// System Resources
	mrc := define.ResourceConfig{
		CPUs:     opts.CPUS,
		DiskSize: strongunits.GiB(opts.DiskSize),
		Memory:   strongunits.MiB(opts.Memory),
	}
	mc.Resources = mrc
	mc.Created = time.Now()
	return mc, nil
}

// loadMachineFromFQPath stub function for loading a JSON configuration file and returning
// a machineconfig.  this should only be called if you know what you are doing.
func loadMachineFromFQPath(path *define.VMFile) (*MachineConfig, error) {
	mc := new(MachineConfig)
	b, err := path.Read()
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(b, mc); err != nil {
		return nil, fmt.Errorf("unable to load machine config file: %q", err)
	}

	return mc, err
}

// LoadMachinesInDir returns all the machineconfigs located in given dir
func LoadMachinesInDir(machinedirs *define.MachineDirs) (map[string]*MachineConfig, error) {
	mcs := make(map[string]*MachineConfig)
	if err := filepath.WalkDir(
		machinedirs.ConfigDir.GetPath(),
		func(path string, d fs.DirEntry, err error) error {
			if strings.HasSuffix(d.Name(), ".json") {
				fullPath, err := machinedirs.ConfigDir.AppendToNewVMFile(d.Name())
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
				mc.Dirs = machinedirs
				mcs[mc.Name] = mc
			}
			return nil
		}); err != nil {
		return nil, err
	}
	return mcs, nil
}
