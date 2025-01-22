//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bauklotze/pkg/machine/defconfig"
	"bauklotze/pkg/machine/vmconfig"
	"fmt"
	"os"
	"runtime"

	"bauklotze/pkg/machine/krunkit"
	"bauklotze/pkg/machine/vfkit"

	"github.com/sirupsen/logrus"
)

// Get current hypervisor provider with default configure
func Get() (vmconfig.VMProvider, error) {
	cfg := defconfig.VMConfig()

	provider := cfg.Provider // provider= ""
	if runtime.GOARCH == "amd64" && runtime.GOOS == "darwin" {
		provider = defconfig.VFkit.String()
	}

	if runtime.GOARCH == "arm64" && runtime.GOOS == "darwin" {
		provider = defconfig.LibKrun.String()
	}

	// OVM_PROVIDER overwrite the provider
	if providerOverride, found := os.LookupEnv("OVM_PROVIDER"); found {
		logrus.Warnf("OVM_PROVIDER is set %s, overriding provider", providerOverride)
		provider = providerOverride
	}

	vmType, err := defconfig.ParseVMType(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vm type: %w", err)
	}

	switch vmType {
	case defconfig.VFkit:
		return new(vfkit.VFkitStubber), nil
	case defconfig.LibKrun:
		return new(krunkit.LibKrunStubber), nil
	default:
		return nil, fmt.Errorf("unsupported virtualization provider: `%s`", vmType.String())
	}
}

func GetAll() []vmconfig.VMProvider {
	return []vmconfig.VMProvider{
		new(krunkit.LibKrunStubber),
		new(vfkit.VFkitStubber),
	}
}
