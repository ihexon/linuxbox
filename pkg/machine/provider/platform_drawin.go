//go:build darwin && arm64

package provider

import (
	"bauklotze/pkg/config"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/krunkit"
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
	resolvedVMType, err := define.ParseVMType(provider, define.LibKrun)
	switch resolvedVMType {
	case define.LibKrun:
		return new(krunkit.LibKrunStubber), nil
	default:
		return nil, fmt.Errorf("unsupported virtualization provider: `%s`", resolvedVMType.String())
	}
}

func GetAll() []vmconfigs.VMProvider {
	return []vmconfigs.VMProvider{
		new(krunkit.LibKrunStubber),
	}
}
