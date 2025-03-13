package backend

import (
	"bauklotze/pkg/machine/ssh/service"
	"net/http"

	"bauklotze/pkg/api/types"
	"bauklotze/pkg/api/utils"
	"bauklotze/pkg/machine/vmconfig"
)

func StopVM(w http.ResponseWriter, r *http.Request) {
	mc := r.Context().Value(types.McKey).(*vmconfig.MachineConfig)
	if mc == nil {
		utils.Error(w, http.StatusInternalServerError, ErrMachineConfigNull)
		return
	}

	// in busybox init system, reboot cause vCPU 0 received shutdown signal so the
	// krunkit will be shutdown after the vm shutdown
	err := service.GracefulShutdownVK(mc)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, ErrStopVMFailed)
	}
}
