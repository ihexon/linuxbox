package vmconfigs

import (
	"bauklotze/pkg/machine/define"
	"github.com/sirupsen/logrus"
	"io/fs"
	"path/filepath"
	"strings"
)

func (mc *MachineConfig) GVProxySocket() (*define.VMFile, error) {
	machineRuntimeDir, err := mc.RuntimeDir()
	if err != nil {
		return nil, err
	}
	return gvProxySocket(mc.Name, machineRuntimeDir)
}

func LoadMachinesInDir(dirs *define.MachineDirs) (map[string]*MachineConfig, error) {
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
