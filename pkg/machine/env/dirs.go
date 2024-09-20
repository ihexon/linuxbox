package env

import (
	"bauklotze/pkg/machine/machineDefine"
	"os"
	"path/filepath"
	"strings"
)

const prefix_str = "donaldtrump"

func GetMachineDirs(vmType machineDefine.VMType) (*machineDefine.MachineDirs, error) {
	vmFiles := []*machineDefine.VMFile{}

	d, err := getRuntimeDir()
	if err != nil {
		return nil, err
	}
	d = filepath.Join(d, machineDefine.MyName)
	rtVMDir := &machineDefine.VMFile{Path: d}
	vmFiles = append(vmFiles, rtVMDir)

	vmconfDir, err := GetVMConfDir(vmType)
	if err != nil {
		return nil, err
	}
	configDir := &machineDefine.VMFile{Path: vmconfDir}
	vmFiles = append(vmFiles, configDir)

	vmdataDir, err := GetVMDataDir(vmType)
	if err != nil {
		return nil, err
	}
	dataDir := &machineDefine.VMFile{Path: vmdataDir}
	vmFiles = append(vmFiles, dataDir)

	imageCacheDir, err := GetImageCacheDir(vmType)
	if err != nil {
		return nil, err
	}
	cacheDir := &machineDefine.VMFile{Path: imageCacheDir}
	vmFiles = append(vmFiles, cacheDir)

	// Resolve VMFile
	for _, vmf := range vmFiles {
		if err = vmf.Abs(); err != nil {
			return nil, err
		}
		err = vmf.CreatePath()
		if err != nil {
			return nil, err
		}
	}

	dirs := &machineDefine.MachineDirs{
		ConfigDir:     configDir,
		DataDir:       dataDir,
		RuntimeDir:    rtVMDir,
		ImageCacheDir: cacheDir,
	}

	return dirs, err
}

func GetImageCacheDir(vmType machineDefine.VMType) (string, error) {
	p, err := GetVMDataDir(vmType)
	if err != nil {
		return "", err
	}
	return filepath.Join(p, "cache"), nil
}

// GetConfigHome return $HOME/.config,
func GetConfigHome() (string, error) {
	homeDir, _ := GetHomePath()
	return filepath.Join(homeDir, ".config"), nil
}

// GetHomePath return $HOME, If the CustomHomeDir is set, then using CustomHomeDir
func GetHomePath() (string, error) {
	if CustomHomeEnv != "" {
		return CustomHomeEnv, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return home, nil
}

// GetVMConfDir return $HOME/.config/oomol/ovm/machine/{wsl,qemu,libkrun,applehv}
func GetVMConfDir(vmType machineDefine.VMType) (string, error) {
	confDirPrefix, err := GetMachineConfDir()
	if err != nil {
		return "", err
	}
	confDir := filepath.Join(confDirPrefix, vmType.String())
	mkdirErr := os.MkdirAll(confDir, 0755)
	return confDir, mkdirErr
}

// GetMachineConfDir return $HOME/.config/oomol/ovm/machine
func GetMachineConfDir() (string, error) {
	// configDirOfMachine ~/.config/
	configDirOfMachine, err := GetConfigHome()
	if err != nil {
		return "", err
	}
	// ~/.config/oomol/ovm/machine/
	configDirOfMachine = filepath.Join(configDirOfMachine, "oomol", "ovm", "machine")
	return configDirOfMachine, nil
}

func WithBugBoxPrefix(name string) string {
	if !strings.HasPrefix(name, prefix_str) {
		name = prefix_str + name
	}
	return name
}

// GetVMDataDir return $HOME/.local/share/oomol/ovm/machine/{wsl,libkrun}
func GetVMDataDir(vmType machineDefine.VMType) (string, error) {
	dataHomePrefix, err := DataDirMachine()
	if err != nil {
		return "", err
	}
	dataDir := filepath.Join(dataHomePrefix, vmType.String())
	mkdirErr := os.MkdirAll(dataDir, 0755)
	return dataDir, mkdirErr
}

// DataDirMachine return $HOME/.local/share/oomol/ovm/machine/
func DataDirMachine() (string, error) {
	data, err := GetDataDirPrefix()
	if err != nil {
		return "", err
	}
	dataDir := filepath.Join(data, "machine")
	return dataDir, nil
}

// GetDataDirPrefix return $HOME/.local/share/oomol/ovm/
func GetDataDirPrefix() (string, error) {
	home, err := GetHomePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "oomol", "ovm"), nil
}

// GetGlobalDataDir return $HOME/.local/share/oomol/ovm/machine/
func GetGlobalDataDir() (string, error) {
	dataDir, err := DataDirMachine()
	if err != nil {
		return "", err
	}
	return dataDir, os.MkdirAll(dataDir, 0755)
}

// GetSSHIdentityPath returns the path to the expected SSH private key
func GetSSHIdentityPath(name string) (string, error) {
	datadir, err := GetGlobalDataDir()
	if err != nil {
		return "", err
	}
	datadir, err = filepath.Abs(datadir)
	return filepath.Join(datadir, name), nil
}
