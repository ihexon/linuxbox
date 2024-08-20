package env

import (
	"bauklotze/pkg/machine/define"
	"os"
	"path/filepath"
	"strings"
)

func GetMachineDirs(vmType define.VMType) (*define.MachineDirs, error) {
	rtDir, err := getRuntimeDir()
	if err != nil {
		return nil, err
	}
	rtDirOfVM := define.VMFile{Path: rtDir}

	confDirOfVM, err := GetConfDirOfVM(vmType)
	if err != nil {
		return nil, err
	}
	configDir := define.VMFile{Path: confDirOfVM}

	dataDirOfVM, err := GetDataHomeOfVM(vmType)
	if err != nil {
		return nil, err
	}
	dataDir := define.VMFile{Path: dataDirOfVM}

	dirs := define.MachineDirs{
		ConfigDir:  &configDir,
		DataDir:    &dataDir,
		RuntimeDir: &rtDirOfVM,
	}
	return &dirs, err
}

func GetTmpDir() (string, error) {
	tmpDir, _ := getRuntimeDir()
	runtimeDir := filepath.Join(tmpDir, "oomol", "ovm")
	return runtimeDir, nil
}

// GetDataHomeOfVM return $HOME/.local/share/oomol/ovm/wsl
func GetDataHomeOfVM(vmType define.VMType) (string, error) {
	dataHomePrefix, err := GetDataHomePrefix()
	if err != nil {
		return "", err
	}
	dataDir := filepath.Join(dataHomePrefix, vmType.String())
	mkdirErr := os.MkdirAll(dataDir, 0755)
	return dataDir, mkdirErr
}

// GetDataHomePrefix return $HOME/.local/share
func GetDataHomePrefix() (string, error) {
	home, err := GetHomePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "oomol", "ovm"), nil
}

// GetConfDirOfVM return $HOME/.config/oomol/ovm/machine/{wsl,qemu}
func GetConfDirOfVM(vmType define.VMType) (string, error) {
	confDirPrefix, err := GetConfigDirOfMachine()
	if err != nil {
		return "", err
	}
	confDir := filepath.Join(confDirPrefix, vmType.String())
	mkdirErr := os.MkdirAll(confDir, 0755)
	return confDir, mkdirErr
}

// GetConfigDirOfMachine return $HOME/.config/oomol/ovm/machine
func GetConfigDirOfMachine() (string, error) {
	configDirOfMachine, err := GetConfigHome()
	if err != nil {
		return "", err
	}
	configDirOfMachine = filepath.Join(configDirOfMachine, "oomol", "ovm", "machine")
	return configDirOfMachine, nil
}

// GetConfigHome return $HOME/.config
func GetConfigHome() (string, error) {
	homeDir, _ := GetHomePath()
	return filepath.Join(homeDir, ".config"), nil
}

// GetHomePath return $HOME
func GetHomePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return home, nil
}

const prefix_str = "donaldtrump"

func WithBugBoxPrefix(name string) string {
	if !strings.HasPrefix(name, prefix_str) {
		name = prefix_str + name
	}
	return name
}
