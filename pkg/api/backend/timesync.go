//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package backend

import (
	"bauklotze/pkg/machine/vmconfig"
	"bauklotze/pkg/ssh"
	"errors"
	"fmt"
	"net/http"
	"time"

	"bauklotze/pkg/api/utils"
	provider2 "bauklotze/pkg/machine/provider"
)

type timeStruct struct {
	Time string `json:"time"`
	Tz   string `json:"tz"`
}

func getCurrentTime() *timeStruct {
	currentTime := time.Now()
	tz, _ := currentTime.Zone()
	return &timeStruct{
		Time: currentTime.Format("2006-01-02 15:04:05"),
		Tz:   tz,
	}
}

func getVMMc(vmName string) (*vmconfig.MachineConfig, error) {
	providers = provider2.GetAll()
	for _, sprovider := range providers {
		dirs, err := vmconfig.GetMachineDirs(sprovider.VMType())
		if err != nil {
			return nil, fmt.Errorf("failed to get machine dirs: %w", err)
		}
		mcs, err := vmconfig.LoadMachinesInDir(dirs)
		if err != nil {
			return nil, fmt.Errorf("failed to load machines in dir: %w", err)
		}
		if mc, exists := mcs[vmName]; exists {
			return mc, nil
		}
	}
	return nil, errors.New("unknown error")
}

func TimeSync(w http.ResponseWriter, r *http.Request) {
	timeSt := getCurrentTime()

	name := utils.GetName(r)
	mc, err := getVMMc(name)

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err)
		return
	}

	if sshError := ssh.CommonSSHSilent(mc.SSH.RemoteUsername, mc.SSH.IdentityPath, mc.VMName, mc.SSH.Port, []string{"date -s " + "'" + timeSt.Time + "'"}); sshError != nil {
		utils.Error(w, http.StatusInternalServerError, sshError)
		return
	}

	utils.WriteResponse(w, http.StatusOK, timeSt)
}
