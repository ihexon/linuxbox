//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package vmconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"bauklotze/pkg/machine/defconfig"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/io"
)

var (
	workSpace *io.VMFile
	Once      sync.Once
)

// SetWorkSpace is a function given a string, it returns a pointer to a VMDir and an error
func SetWorkSpace(s string) (*io.VMFile, error) {
	var err error
	Once.Do(func() {
		workSpace, err = io.NewMachineFile(s)
	})
	if err != nil {
		return nil, fmt.Errorf("%w, %w", define.ErrConstructVMFile, err)
	}
	return workSpace, nil
}

// GetWorkSpace is a function that returns the workspace and an error
func GetWorkSpace() (*io.VMFile, error) {
	if workSpace == nil || workSpace.GetPath() == "" {
		return nil, fmt.Errorf("workspace is not set")
	}
	return workSpace, nil
}

var Dirs MachineDirs

func GetMachineDirs(vmType defconfig.VMType) (*MachineDirs, error) {
	d, err := GetWorkSpace()
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	socksDir, err := io.NewMachineFile(filepath.Join(d.GetPath(), define.SocksPrefixDir, vmType.String()))
	if err != nil {
		return nil, fmt.Errorf("unable to new machine file in %s: %w", socksDir, err)
	}

	dataDir, err := io.NewMachineFile(filepath.Join(d.GetPath(), define.DataPrefixDir, vmType.String()))
	if err != nil {
		return nil, fmt.Errorf("unable to new machine file in %s: %w", dataDir, err)
	}

	configDir, err := io.NewMachineFile(filepath.Join(d.GetPath(), define.ConfigPrefixDir, vmType.String()))
	if err != nil {
		return nil, fmt.Errorf("unable to new machine file in %s: %w", configDir, err)
	}

	logsDir, err := io.NewMachineFile(filepath.Join(d.GetPath(), define.LogPrefixDir, vmType.String()))
	if err != nil {
		return nil, fmt.Errorf("unable to new machine file in %s: %w", logsDir, err)
	}

	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("unable to get executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return nil, fmt.Errorf("unable to eval symlinks: %w", err)
	}

	binDir := filepath.Dir(execPath)                                                                   // $BINDIR
	libexecDir, err := io.NewMachineFile(filepath.Join(filepath.Dir(binDir), define.LibexecPrefixDir)) // $BINDIR/../libexec
	if err != nil {
		return nil, fmt.Errorf("unable to new machine file in %s: %w", logsDir, err)
	}

	var hypervisorBin *io.VMFile
	// If hypervisor is Krunkit, use krunkit binary name
	if vmType.String() == defconfig.LibKrun.String() {
		hypervisorBin = &io.VMFile{Path: filepath.Join(libexecDir.GetPath(), define.KrunkitBinaryName)}
	} else {
		// If hypervisor is vfkit, use vfkit binary name
		hypervisorBin = &io.VMFile{Path: filepath.Join(libexecDir.GetPath(), define.VfkitBinaryName)}
	}

	networkProviderBin := &io.VMFile{
		Path: filepath.Join(libexecDir.GetPath(), define.GvProxyBinaryName),
	}

	Dirs = MachineDirs{
		ConfigDir: configDir,
		DataDir:   dataDir,
		SocksDir:  socksDir,
		LogsDir:   logsDir,
		Hypervisor: &Hypervisor{
			Bin:     hypervisorBin,
			LibsDir: libexecDir,
		},
		NetworkProvider: &NetworkProvider{
			LibsDir: libexecDir,
			Bin:     networkProviderBin,
		},
	}

	if err = os.MkdirAll(socksDir.GetPath(), 0755); err != nil {
		return nil, fmt.Errorf("unable to create runtime dir: %s: %w", socksDir.GetPath(), err)
	}
	if err = os.MkdirAll(configDir.GetPath(), 0755); err != nil {
		return nil, fmt.Errorf("unable to create config dir: %s: %w", configDir.GetPath(), err)
	}
	if err = os.MkdirAll(logsDir.GetPath(), 0755); err != nil {
		return nil, fmt.Errorf("unable to create logs dir: %s: %w", logsDir.GetPath(), err)
	}

	return &Dirs, nil
}

// GetSSHIdentityPath returns the path to the expected SSH private key
func GetSSHIdentityPath(vmType defconfig.VMType) (*io.VMFile, error) {
	dirs, err := GetMachineDirs(vmType)
	if err != nil {
		return nil, fmt.Errorf("failed to get machine dirs: %w", err)
	}

	sshkey, err := io.NewMachineFile(filepath.Join(dirs.DataDir.GetPath(), define.DefaultIdentityName))
	if err != nil {
		return nil, fmt.Errorf("unable to new machine file in %s: %w", dirs.DataDir.GetPath(), err)
	}

	return sshkey, nil
}
