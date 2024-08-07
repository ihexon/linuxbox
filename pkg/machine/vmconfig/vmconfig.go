package vmconfig

import (
	"bauklotze/pkg/machine/define"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

type MachineConfig define.MachineConfig

func (mc *MachineConfig) RemoveRuntimeFiles() ([]string, func() error, error) {
	return nil, nil, nil
}

func NewMachineConfig(opts define.InitOptions, dirs *define.MachineDirs, vmtype define.VMType) (*define.MachineConfig, error) {
	mc := new(define.MachineConfig)
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
		DiskSize: define.GiB(opts.DiskSize),
		Memory:   define.MiB(opts.Memory),
	}
	mc.Resources = mrc
	mc.Created = time.Now()
	return mc, nil
}

// loadMachineFromFQPath stub function for loading a JSON configuration file and returning
// a machineconfig.  this should only be called if you know what you are doing.
func loadMachineFromFQPath(path *define.VMFile) (*define.MachineConfig, error) {
	mc := new(define.MachineConfig)
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
func LoadMachinesInDir(machinedirs *define.MachineDirs) (map[string]*define.MachineConfig, error) {
	mcs := make(map[string]*define.MachineConfig)
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
