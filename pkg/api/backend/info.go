//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package backend

import (
	"bauklotze/pkg/machine/vmconfig"
	"net/http"

	"bauklotze/pkg/api/utils"
	provider2 "bauklotze/pkg/machine/provider"
)

func getPodmanConnection(vmName string) *vmconfig.MachineConfig {
	providers = provider2.GetAll()
	for _, s := range providers {
		dirs, err := vmconfig.GetMachineDirs(s.VMType())
		if err != nil {
			return nil
		}
		mcs, err := vmconfig.LoadMachinesInDir(dirs)
		if err != nil {
			return nil
		}

		for name, mc := range mcs {
			if name == vmName {
				return mc
			}
		}
	}
	return nil
}

func GetInfos(w http.ResponseWriter, r *http.Request) {
	name := utils.GetName(r)
	mc := getPodmanConnection(name)
	utils.WriteResponse(w, http.StatusOK, mc)
}
