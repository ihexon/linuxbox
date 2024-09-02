//go:build darwin && arm64

package provider

import (
	"bauklotze/pkg/config"
	"bauklotze/pkg/machine/krunkit"
	"bauklotze/pkg/machine/machineDefine"
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
	resolvedVMType, err := machineDefine.ParseVMType(provider, machineDefine.LibKrun)
	switch resolvedVMType {
	case machineDefine.LibKrun:
		return new(krunkit.LibKrunStubber), nil
	default:
		return nil, fmt.Errorf("unsupported virtualization provider: `%s`", resolvedVMType.String())
	}
}

func GetAll(_ bool) ([]vmconfigs.VMProvider, error) {
	return []vmconfigs.VMProvider{
		new(krunkit.LibKrunStubber),
	}, nil
}
