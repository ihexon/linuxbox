//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package provider

import (
	"fmt"
	"os"
	"runtime"

	"bauklotze/pkg/config"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/krunkit"
	"bauklotze/pkg/machine/vfkit"
	"bauklotze/pkg/machine/vmconfigs"

	"github.com/sirupsen/logrus"
)

// Get current hypervisor provider with default configure
func Get() (vmconfigs.VMProvider, error) {
	cfg := config.Default()

	provider := cfg.Machine.Provider // provider= ""
	if runtime.GOARCH == "amd64" && runtime.GOOS == "darwin" {
		provider = define.VFkit.String()
	}

	if runtime.GOARCH == "arm64" && runtime.GOOS == "darwin" {
		provider = define.LibKrun.String()
	}

	// OVM_PROVIDER overwrite the provider
	if providerOverride, found := os.LookupEnv("OVM_PROVIDER"); found {
		logrus.Warnf("OVM_PROVIDER is set %s, overriding provider", providerOverride)
		provider = providerOverride
	}

	resolvedVMType, err := define.ParseVMType(provider, define.LibKrun)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vm type: %w", err)
	}

	switch resolvedVMType {
	case define.VFkit:
		return new(vfkit.VFkitStubber), nil
	case define.LibKrun:
		return new(krunkit.LibKrunStubber), nil
	default:
		return nil, fmt.Errorf("unsupported virtualization provider: `%s`", resolvedVMType.String())
	}
}

func GetAll() []vmconfigs.VMProvider {
	return []vmconfigs.VMProvider{
		new(krunkit.LibKrunStubber),
		new(vfkit.VFkitStubber),
	}
}
