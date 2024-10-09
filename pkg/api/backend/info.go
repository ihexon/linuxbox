package backend

import (
	"bauklotze/pkg/api/utils"
	"bauklotze/pkg/machine/env"
	provider2 "bauklotze/pkg/machine/provider"
	"bauklotze/pkg/machine/vmconfigs"
	"net/http"
)

func getPodmanConnection(vmName string) *vmconfigs.MachineConfig {
	providers = provider2.GetAll()
	for _, s := range providers {
		dirs, err := env.GetMachineDirs(s.VMType())
		if err != nil {
			return nil
		}
		mcs, err := vmconfigs.LoadMachinesInDir(dirs)
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
