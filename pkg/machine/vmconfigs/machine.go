package vmconfigs

import (
	"bauklotze/pkg/machine/machineDefine"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func (mc *MachineConfig) GVProxySocket() (*machineDefine.VMFile, error) {
	machineRuntimeDir, err := mc.RuntimeDir()
	if err != nil {
		return nil, err
	}
	return gvProxySocket(mc.Name, machineRuntimeDir)
}

func (mc *MachineConfig) APISocket() (*machineDefine.VMFile, error) {
	machineRuntimeDir, err := mc.RuntimeDir()
	if err != nil {
		return nil, err
	}
	return apiSocket(mc.Name, machineRuntimeDir)
}

func apiSocket(name string, socketDir *machineDefine.VMFile) (*machineDefine.VMFile, error) {
	socketName := name + "-api.sock"
	return socketDir.AppendToNewVMFile(socketName)
}

func (mc *MachineConfig) Lock() {
	mc.lock.Lock()
}

// Unlock removes an existing lock
func (mc *MachineConfig) Unlock() {
	mc.lock.Unlock()
}

// Refresh reloads the config file from disk
func (mc *MachineConfig) Refresh() error {
	content, err := os.ReadFile(mc.ConfigPath.GetPath())
	if err != nil {
		return err
	}
	return json.Unmarshal(content, mc)
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
