//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package backend

import (
	"net/http"

	"bauklotze/pkg/api/types"
	"bauklotze/pkg/machine/vmconfig"

	"bauklotze/pkg/api/utils"
)

type Resp struct {
	PodmanSocketPath string `json:"podmanSocketPath"`
	SSHPort          int    `json:"sshPort"`
	SSHUser          string `json:"sshUser"`
	HostEndpoint     string `json:"hostEndpoint"`
}

// GetInfos Get machine configures
func GetInfos(w http.ResponseWriter, r *http.Request) {
	mc := r.Context().Value(types.McKey).(*vmconfig.MachineConfig)
	if mc == nil {
		utils.Error(w, http.StatusInternalServerError, ErrMachineConfigNull)
		return
	}

	utils.WriteResponse(w, http.StatusOK, &Resp{
		PodmanSocketPath: mc.GvProxy.HostSocks[0],
		SSHPort:          mc.SSH.Port,
		SSHUser:          mc.SSH.RemoteUsername,
		HostEndpoint:     "host.containers.internal",
	})
}
