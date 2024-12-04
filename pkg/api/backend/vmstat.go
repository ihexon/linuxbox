//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package backend

import (
	"errors"
	"net/http"

	"bauklotze/pkg/api/utils"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	provider2 "bauklotze/pkg/machine/provider"
	"bauklotze/pkg/machine/vmconfigs"
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
		return "Stopped"
	default:
	}
	return "Stopped"
}

func getVMstat(vmName string) (statType, error) {
	providers = provider2.GetAll()
	for _, sprovider := range providers {
		dirs, err := env.GetMachineDirs(sprovider.VMType())
		if err != nil {
			return unknown, err
		}
		mcs, err := vmconfigs.LoadMachinesInDir(dirs)
		if err != nil {
			return unknown, err
		}
		if mc, exists := mcs[vmName]; exists {
			state, err := sprovider.State(mc)
			if err != nil {
				return unknown, err
			}
			switch state {
			case define.Running:
				return running, nil
			case define.Stopped:
				return stopped, nil
			}
		}
	}
	return unknown, errors.New("unknown state")
}

func GetVMStat(w http.ResponseWriter, r *http.Request) {
	s := &vmStat{
		CurrentStat: stopped.String(),
	}

	name := utils.GetName(r)
	stat, err := getVMstat(name)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err)
		return
	}

	s.VMName = name
	s.CurrentStat = stat.String()

	utils.WriteResponse(w, http.StatusOK, &vmStat{
		VMName:      s.VMName,
		CurrentStat: s.CurrentStat,
	})
}
