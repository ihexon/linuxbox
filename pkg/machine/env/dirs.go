package env

import (
	"bauklotze/pkg/machine/define"
	"os"
	"path/filepath"
	"strings"
)

const prefix_str = "donaldtrump"

func GetMachineDirs(vmType define.VMType) (*define.MachineDirs, error) {
	tmpDir, err := getTMPDir()
	if err != nil {
		return nil, err
	}
	tmpDir = filepath.Join(tmpDir, "ovm")
	ovmTMPDir := define.VMFile{Path: tmpDir}

	vmconfDir, err := GetVMConfDir(vmType)
	if err != nil {
		return nil, err
	}
	configDir := define.VMFile{Path: vmconfDir}

	vmdataDir, err := GetVMDataDir(vmType)
	if err != nil {
		return nil, err
	}
	dataDir := define.VMFile{Path: vmdataDir}

	dirs := define.MachineDirs{
		ConfigDir:  &configDir,
		DataDir:    &dataDir,
		RuntimeDir: &ovmTMPDir,
	}
	return &dirs, err
}

// GetConfigHome return $HOME/.config
func getConfigHome() (string, error) {
	homeDir, _ := getHomePath()
	return filepath.Join(homeDir, ".config"), nil
}

// GetVMConfDir return $HOME/.config/oomol/ovm/machine/{wsl,qemu,libkrun,applehv}
func GetVMConfDir(vmType define.VMType) (string, error) {
	confDirPrefix, err := getMachineDir()
	if err != nil {
		return "", err
	}
	confDir := filepath.Join(confDirPrefix, vmType.String())
	mkdirErr := os.MkdirAll(confDir, 0755)
	return confDir, mkdirErr
}

// getMachineDir return $HOME/.config/oomol/ovm/machine
func getMachineDir() (string, error) {
	// configDirOfMachine ~/.config/
	configDirOfMachine, err := getConfigHome()
	if err != nil {
		return "", err
	}
	// ~/.config/oomol/ovm/machine/
	configDirOfMachine = filepath.Join(configDirOfMachine, "oomol", "ovm", "machine")
	return configDirOfMachine, nil
}

// getHomePath return $HOME
func getHomePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return home, nil
}

func WithBugBoxPrefix(name string) string {
	if !strings.HasPrefix(name, prefix_str) {
		name = prefix_str + name
	}
	return name
}

// GetVMDataDir return $HOME/.local/share/oomol/ovm/machine/wsl
func GetVMDataDir(vmType define.VMType) (string, error) {
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
	home, err := getHomePath()
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
	return filepath.Join(datadir, name), nil
}
