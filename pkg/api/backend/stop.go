package backend

import (
	"net/http"

	"bauklotze/pkg/machine/ssh/service"

	"bauklotze/pkg/api/types"
	"bauklotze/pkg/api/utils"
	"bauklotze/pkg/machine/vmconfig"

	"github.com/sirupsen/logrus"
)

func StopVM(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("Request /stop")
	mc := r.Context().Value(types.McKey).(*vmconfig.MachineConfig)
	if mc == nil {
		utils.Error(w, http.StatusInternalServerError, ErrMachineConfigNull)
		return
	}

	if err := service.GracefulShutdownVK(mc); err != nil {
		utils.Error(w, http.StatusInternalServerError, ErrStopVMFailed)
	}
}
