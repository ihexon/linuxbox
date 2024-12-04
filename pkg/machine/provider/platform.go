//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/vmconfigs"
)

// GetAllMachinesAndRootfulness collects all podman machine configs and returns
// a map in the format: { machineName: isRootful }
func GetAllMachinesAndRootfulness() (map[string]bool, error) {
	providers := GetAll()
	machines := map[string]bool{}
	for _, provider := range providers {
		dirs, err := env.GetMachineDirs(provider.VMType())
		if err != nil {
			return nil, err
		}
		providerMachines, err := vmconfigs.LoadMachinesInDir(dirs)
		if err != nil {
			return nil, err
		}

		for n, m := range providerMachines {
			machines[n] = m.HostUser.Rootful
		}
	}

	return machines, nil
}
