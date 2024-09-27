package env

import (
	"bauklotze/pkg/machine/define"
	"os"
	"path/filepath"
)

// GetBauklotzeHomePath return $HOME/.Bauklotze_dir.
// If the CustomHomeDir is set, then using CustomHomeDir/
func GetBauklotzeHomePath() (string, error) {
	if CustomHomeEnv != "" {
		return CustomHomeEnv, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, define.WorkDir), nil
}

// ConfDirPrefix return ${BauklotzeHomePath}/config,
func ConfDirPrefix() (string, error) {
	homeDir, _ := GetBauklotzeHomePath()
	return filepath.Join(homeDir, "config"), nil // ${BauklotzeHomePath}/config
}

// GetConfDir ${BauklotzeHomePath}/config/{wsl,libkrun,qemu,hyper...}
func GetConfDir(vmType define.VMType) (string, error) {
	confDirPrefix, err := ConfDirPrefix() // ${BauklotzeHomePath}/config
	if err != nil {
		return "", err
	}
	confDir := filepath.Join(confDirPrefix, vmType.String())
	mkdirErr := os.MkdirAll(confDir, 0755)
	return confDir, mkdirErr // ${BauklotzeHomePath}/config/wsl2
}

// DataDirPrefix returns the path prefix for all machine data files
func DataDirPrefix() (string, error) {
	d, err := GetBauklotzeHomePath() // ${BauklotzeHomePath}
	if err != nil {
		return "", err
	}
	dataDir := filepath.Join(d, "data")
	return dataDir, nil // ${BauklotzeHomePath}/data
}

// GetDataDir ${BauklotzeHomePath}/data/{wsl2,libkrun,qemu,hyper...}
func GetDataDir(vmType define.VMType) (string, error) {
	dataDirPrefix, err := DataDirPrefix() // ${BauklotzeHomePath}/data
	if err != nil {
		return "", err
	}
	dataDir := filepath.Join(dataDirPrefix, vmType.String())
	mkdirErr := os.MkdirAll(dataDir, 0755)
	return dataDir, mkdirErr // ${BauklotzeHomePath}/data/{wsl2,libkrun,qemu,hyper...}
}

func GetGlobalDataDir() (string, error) {
	dataDir, err := DataDirPrefix()
	if err != nil {
		return "", err
	}
	return dataDir, os.MkdirAll(dataDir, 0755)
}

func GetMachineDirs(vmType define.VMType) (*define.MachineDirs, error) {
	rtDir, err := getRuntimeDir()
	if err != nil {
		return nil, err
	}
	rtDirFile, err := define.NewMachineFile(rtDir, nil)

	dataDir, err := GetDataDir(vmType)
	dataDirFile, err := define.NewMachineFile(dataDir, nil)
	imageCacheDir, err := dataDirFile.AppendToNewVMFile("cache", nil)

	configDir, err := GetConfDir(vmType)
	if err != nil {
		return nil, err
	}
	configDirFile, err := define.NewMachineFile(configDir, nil)

	dirs := define.MachineDirs{
		ConfigDir:     configDirFile, // ${BauklotzeHomePath}/config/{wsl,libkrun,qemu,hyper...}
		DataDir:       dataDirFile,   // ${BauklotzeHomePath}/data/{wsl2,libkrun,qemu,hyper...}
		ImageCacheDir: imageCacheDir, // ${BauklotzeHomePath}/data/{wsl2,libkrun,qemu,hyper...}/cache
		RuntimeDir:    rtDirFile,     // ${BauklotzeHomePath}/tmp/
	}
	if err = os.MkdirAll(rtDir, 0755); err != nil {
		return nil, err
	}
	if err = os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}
	err = os.MkdirAll(imageCacheDir.GetPath(), 0755)
	return &dirs, err
}

// GetSSHIdentityPath returns the path to the expected SSH private key
func GetSSHIdentityPath(name string) (string, error) {
	datadir, err := GetGlobalDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(datadir, name), nil
}
