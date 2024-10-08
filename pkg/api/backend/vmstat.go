package backend

import (
	"bauklotze/pkg/api/utils"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	provider2 "bauklotze/pkg/machine/provider"
	"bauklotze/pkg/machine/vmconfigs"
	"net/http"
)

type statType int

type vmStat struct {
	VMName      string
	CurrentStat string
}

const (
	stopped statType = iota
	running
	unknown
)

func (v statType) String() string {
	switch v {
	case stopped:
		return "Stopped"
	case running:
		return "Running"
	case unknown:
		return "Unknown"
	default:
	}
	return "Unknown"
}

func getVMstat(vmName string) statType {
	providers = provider2.GetAll()
	for _, sprovider := range providers {
		dirs, err := env.GetMachineDirs(sprovider.VMType())
		if err != nil {
			return unknown
		}
		mcs, err := vmconfigs.LoadMachinesInDir(dirs)
		if err != nil {
			return unknown
		}
		for name, mc := range mcs {
			if name == vmName {
				state, _ := sprovider.State(mc)
				if state == define.Running {
					return running
				} else {
					return stopped
				}
			}
		}
	}
	return unknown
}

func GetVMStat(w http.ResponseWriter, r *http.Request) {
	s := &vmStat{
		CurrentStat: stopped.String(),
	}

	name := utils.GetName(r)
	stat := getVMstat(name)
	s.CurrentStat = stat.String()
	s.VMName = name

	utils.WriteResponse(w, http.StatusOK, s.CurrentStat)
}
