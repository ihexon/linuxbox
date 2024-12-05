//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package env

import (
	"fmt"
	"os"
	"path/filepath"

	"bauklotze/pkg/machine/define"
)

func GetBauklotzeHomePath() (string, error) {
	home := os.Getenv(BauklotzeHome)
	if home == "" {
		return "", fmt.Errorf("%s is not set", BauklotzeHome)
	}
	return home, nil
}

func ConfDirPrefix() (string, error) {
	homeDir, err := GetBauklotzeHomePath()
	if err != nil {
		return "", fmt.Errorf("unable to get home path: %w", err)
	}
	return filepath.Join(homeDir, "config"), nil
}

func GetLogsDir() (string, error) {
	homeDir, err := GetBauklotzeHomePath()
	if err != nil {
		return "", fmt.Errorf("unable to get home path: %w", err)
	}
	return filepath.Join(homeDir, "logs"), nil
}

func GetConfDir(vmType define.VMType) (string, error) {
	confDirPrefix, err := ConfDirPrefix()
	if err != nil {
		return "", err
	}
	confDir := filepath.Join(confDirPrefix, vmType.String())
	if err := os.MkdirAll(confDir, 0755); err != nil {
		return "", fmt.Errorf("unable to create conf dir: %w", err)
	}
	return confDir, nil
}

// DataDirPrefix returns the path prefix for all machine data files
func DataDirPrefix() (string, error) {
	d, err := GetBauklotzeHomePath()
	if err != nil {
		return "", fmt.Errorf("unable to get home path: %w", err)
	}
	dataDir := filepath.Join(d, "data")
	return dataDir, nil
}

func GetDataDir(vmType define.VMType) (string, error) {
	dataDirPrefix, err := DataDirPrefix()
	if err != nil {
		return "", err
	}
	dataDir := filepath.Join(dataDirPrefix, vmType.String())
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", fmt.Errorf("unable to create data dir: %w", err)
	}

	return dataDir, nil
}

func GetGlobalDataDir() (string, error) {
	dataDir, err := DataDirPrefix()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", fmt.Errorf("unable to create global data dir: %w", err)
	}

	return dataDir, nil
}

func GetMachineDirs(vmType define.VMType) (*define.MachineDirs, error) {
	rtDir, err := getRuntimeDir()
	if err != nil {
		return nil, err
	}

	rtDirFile, err := define.NewMachineFile(rtDir, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to new machine file in %s: %w", rtDir, err)
	}

	dataDir, err := GetDataDir(vmType)
	if err != nil {
		return nil, err
	}

	dataDirFile, err := define.NewMachineFile(dataDir, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to new machine file in %s: %w", dataDir, err)
	}

	imageCacheDir, err := dataDirFile.AppendToNewVMFile("cache", nil)
	if err != nil {
		return nil, fmt.Errorf("unable to append cache to new vm file in %s: %w", dataDir, err)
	}

	configDir, err := GetConfDir(vmType)
	if err != nil {
		return nil, err
	}
	configDirFile, err := define.NewMachineFile(configDir, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to new machine file in %s: %w", configDir, err)
	}

	logsDir, err := GetLogsDir()
	if err != nil {
		return nil, err
	}
	logsDirVMFile, err := define.NewMachineFile(logsDir, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to new machine file in %s: %w", logsDir, err)
	}

	dirs := define.MachineDirs{
		ConfigDir:     configDirFile, // ${BauklotzeHomePath}/config/{wsl,libkrun,qemu,hyper...}
		DataDir:       dataDirFile,   // ${BauklotzeHomePath}/data/{wsl2,libkrun,qemu,hyper...}
		ImageCacheDir: imageCacheDir, // ${BauklotzeHomePath}/data/{wsl2,libkrun,qemu,hyper...}/cache
		RuntimeDir:    rtDirFile,     // ${BauklotzeHomePath}/tmp/
		LogsDir:       logsDirVMFile, // ${BauklotzeHomePath}/logs
	}
	if err = os.MkdirAll(rtDir, 0755); err != nil {
		return nil, fmt.Errorf("unable to create runtime dir: %s: %w", rtDir, err)
	}
	if err = os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("unable to create config dir: %s: %w", configDir, err)
	}
	if err = os.MkdirAll(logsDirVMFile.GetPath(), 0755); err != nil {
		return nil, fmt.Errorf("unable to create logs dir: %s: %w", logsDirVMFile.GetPath(), err)
	}
	err = os.MkdirAll(imageCacheDir.GetPath(), 0755)
	return &dirs, fmt.Errorf("unable to create image cache dir: %s: %w", imageCacheDir.GetPath(), err)
}

// GetSSHIdentityPath returns the path to the expected SSH private key
func GetSSHIdentityPath(name string) (string, error) {
	datadir, err := GetGlobalDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(datadir, name), nil
}
