//go:build darwin && !linux && !windows

package provider

import (
	"bauklotze/pkg/config"
	"bauklotze/pkg/machine/apple"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfigs"
	"fmt"
	"os"
)

// Get current hypervisor provider with default configure
func Get() (vmconfigs.VMProvider, error) {
	cfg, err := config.Default()
	if err != nil {
		return nil, err
	}
	provider := cfg.Machine.Provider
	// OVM_PROVIDER overwrite the provider
	if providerOverride, found := os.LookupEnv("OVM_PROVIDER"); found {
		provider = providerOverride
	}
	resolvedVMType, err := define.ParseVMType(provider, define.AppleHvVirt)
	switch resolvedVMType {
	case define.AppleHvVirt:
		return new(apple.AppleHVStubber), nil
	default:
		return nil, fmt.Errorf("unsupported virtualization provider: `%s`", resolvedVMType.String())
	}
}

func GetAll(_ bool) ([]vmconfigs.VMProvider, error) {
	return []vmconfigs.VMProvider{
		new(apple.AppleHVStubber),
	}, nil
}
