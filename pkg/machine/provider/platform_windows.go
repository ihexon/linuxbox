//go:build windows && !darwin && !linux

package provider

import (
	"bauklotze/pkg/machine/machineDefine"
	"bauklotze/pkg/machine/wsl"
	"fmt"
	"github.com/sirupsen/logrus"
)

// GetAll get all VMProvider of current platform, windows using wsl.WSLStubber
func GetAll() []machineDefine.VMProvider {
	providers := []machineDefine.VMProvider{
		// Windows only support wsl
		new(wsl.WSLStubber),
	}
	return providers
}

// Get get a provider via configure file
func Get() (machineDefine.VMProvider, error) {
	provider := ""
	// for now autoconfigure provider, but in future I't should read from configure file
	// provider := cfg.Machine.Provider

	resolvedVMType, err := machineDefine.ParseVMType(provider, machineDefine.WSLVirt)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("Using Podman machine with `%s` virtualization provider", resolvedVMType.String())
	switch resolvedVMType {
	case machineDefine.WSLVirt:
		return new(wsl.WSLStubber), nil
	default:
		return nil, fmt.Errorf("unsupported virtualization provider: `%s`", resolvedVMType.String())
	}
}
