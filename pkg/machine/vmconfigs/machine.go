package vmconfigs

import (
	"bauklotze/pkg/machine/define"
	"encoding/json"
	"errors"
	"os"
)

func (mc *MachineConfig) GVProxySocket() (*define.VMFile, error) {
	machineRuntimeDir, err := mc.RuntimeDir()
	if err != nil {
		return nil, err
	}
	return gvProxySocket(mc.Name, machineRuntimeDir)
}

func (mc *MachineConfig) PodmanApiSocketHost() (*define.VMFile, error) {
	machineRuntimeDir, err := mc.RuntimeDir()
	if err != nil {
		return nil, err
	}
	return podmanApiSocketOnHost(mc.Name, machineRuntimeDir)
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

// ConfigDir is a simple helper to obtain the machine config dir
func (mc *MachineConfig) ConfigDir() (*define.VMFile, error) {
	if mc.Dirs == nil || mc.Dirs.ConfigDir == nil {
		return nil, errors.New("no configuration directory set")
	}
	return mc.Dirs.ConfigDir, nil
}
