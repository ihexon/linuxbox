//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package vmconfig

import (
	"fmt"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
)

// GetProvider current hypervisor provider with default configure
func GetProvider() (VMType, error) {
	provider := LibKrun.String()

	if runtime.GOARCH == "amd64" && runtime.GOOS == "darwin" {
		provider = VFkit.String()
	}

	// OVM_PROVIDER overwrite the provider
	if providerOverride, found := os.LookupEnv("OVM_PROVIDER"); found {
		logrus.Warnf("find env OVM_PROVIDER set to %q, overriding provider", providerOverride)
		provider = providerOverride
	}

	vmType, err := ParseVMType(provider)
	if err != nil {
		return Unknown, fmt.Errorf("failed to parse vm type: %w", err)
	}

	return vmType, nil
}
