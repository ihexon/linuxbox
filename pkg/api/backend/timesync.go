package backend

import (
	"bauklotze/pkg/api/utils"
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/env"
	provider2 "bauklotze/pkg/machine/provider"
	"bauklotze/pkg/machine/vmconfigs"
	"errors"
	"net/http"
	"time"
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

func getVMMc(vmName string) (*vmconfigs.MachineConfig, error) {
	providers = provider2.GetAll()
	for _, sprovider := range providers {
		dirs, err := env.GetMachineDirs(sprovider.VMType())
		if err != nil {
			return nil, err
		}
		mcs, err := vmconfigs.LoadMachinesInDir(dirs)
		if err != nil {
			return nil, err
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

	if sshError := machine.CommonSSHSilent(mc.SSH.RemoteUsername, mc.SSH.IdentityPath, mc.Name, mc.SSH.Port, []string{"sudo date -s " + "'" + timeSt.Time + "'"}); sshError != nil {
		utils.Error(w, http.StatusInternalServerError, sshError)
		return
	}

	utils.WriteResponse(w, http.StatusOK, timeSt)

}
