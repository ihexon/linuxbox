//go:build windows && !darwin && !linux

package provider

import (
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/machine/wsl"
	"fmt"
	"github.com/sirupsen/logrus"
)

// GetAll get all VMProvider of current platform, windows using wsl.WSLStubber
func GetAll() []vmconfigs.VMProvider {
	providers := []vmconfigs.VMProvider{
		// Windows only support wsl for now
		new(wsl.WSLStubber),
	}
	return providers
}

func Get() (vmconfigs.VMProvider, error) {
	// HyperVisor with High priority
	provider := ""
	resolvedVMType, err := define.ParseVMType(provider, define.WSLVirt)
	if err != nil {
		return nil, err
	}

	logrus.Infof("Init machine with `%s` virtualization provider", resolvedVMType.String())
	switch resolvedVMType {
	case define.WSLVirt:
		return new(wsl.WSLStubber), nil
	default:
		return nil, fmt.Errorf("unsupported virtualization provider: `%s`", resolvedVMType.String())
	}
}
