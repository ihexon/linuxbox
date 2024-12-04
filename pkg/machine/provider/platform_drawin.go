//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin && arm64

package provider

import (
	"fmt"
	"os"

	"bauklotze/pkg/config"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/krunkit"
	"bauklotze/pkg/machine/vmconfigs"
)

// Get current hypervisor provider with default configure
func Get() (vmconfigs.VMProvider, error) {
	cfg := config.Default()

	provider := cfg.Machine.Provider
	// OVM_PROVIDER overwrite the provider
	if providerOverride, found := os.LookupEnv("OVM_PROVIDER"); found {
		provider = providerOverride
	}
	resolvedVMType, err := define.ParseVMType(provider, define.LibKrun)
	if err != nil {
		return nil, err
	}

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
