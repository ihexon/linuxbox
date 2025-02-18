//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package backend

import (
	"net/http"

	"bauklotze/pkg/api/types"
	"bauklotze/pkg/machine/vmconfig"

	"bauklotze/pkg/api/utils"
)

// GetInfos Get machine configures
func GetInfos(w http.ResponseWriter, r *http.Request) {
	mc := r.Context().Value(types.McKey).(*vmconfig.MachineConfig)
	if mc == nil {
		utils.Error(w, http.StatusInternalServerError, ErrMachineConfigNull)
		return
	}
	utils.WriteResponse(w, http.StatusOK, mc)
}
