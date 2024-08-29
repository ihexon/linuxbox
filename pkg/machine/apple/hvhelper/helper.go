package hvhelper

import (
	"bauklotze/pkg/machine/define"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	vfkit_config "github.com/crc-org/vfkit/pkg/config"
	rest "github.com/crc-org/vfkit/pkg/rest/define"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"io"
	"net/http"
	"time"
)

const (
	inspect = "/vm/inspect"
	state   = "/vm/state"
	version = "/version"
)

const (
	// Values that the machine can be in
	// "VirtualMachineStateStoppedVirtualMachineStateRunningVirtualMachineStatePausedVirtualMachineStateErrorVirtualMachineStateStartingVirtualMachineStatePausingVirtualMachineStateResumingVirtualMachineStateStopping"
	VZMachineStateStopped  VZMachineState = "VirtualMachineStateStopped"
	VZMachineStateRunning  VZMachineState = "VirtualMachineStateRunning"
	VZMachineStatePaused   VZMachineState = "VirtualMachineStatePaused"
	VZMachineStateError    VZMachineState = "VirtualMachineStateError"
	VZMachineStateStarting VZMachineState = "VirtualMachineStateStarting"
	VZMachineStatePausing  VZMachineState = "VirtualMachineStatePausing"
	VZMachineStateResuming VZMachineState = "VirtualMachineStateResuming"
	VZMachineStateStopping VZMachineState = "VirtualMachineStateStopping"
)

type VZMachineState string
type Endpoint string

// Helper describes the use of hvhelper: cmdline and endpoint
type Helper struct {
	LogLevel       logrus.Level
	Endpoint       string
	BinaryPath     *define.VMFile
	VirtualMachine *vfkit_config.VirtualMachine
	Rosetta        bool
}

func (vf *Helper) get(endpoint string, payload io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, endpoint, payload)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

func (vf *Helper) post(endpoint string, payload io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, endpoint, payload)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

func (vf *Helper) stateChange(newState rest.StateChange) error {
	b, err := json.Marshal(rest.VMState{State: string(newState)})
	if err != nil {
		return err
	}
	payload := bytes.NewReader(b)
	serverResponse, err := vf.post(vf.Endpoint+state, payload)
	_ = serverResponse.Body.Close()
	return err
}

func ToMachineStatus(val string) (define.Status, error) {
	switch val {
	case string(VZMachineStateRunning), string(VZMachineStatePausing), string(VZMachineStateResuming), string(VZMachineStateStopping), string(VZMachineStatePaused):
		return define.Running, nil
	case string(VZMachineStateStopped):
		return define.Stopped, nil
	case string(VZMachineStateStarting):
		return define.Starting, nil
	case string(VZMachineStateError):
		return "", errors.New("machine is in error state")
	}
	return "", fmt.Errorf("unknown machine state: %s", val)
}

// getRawState asks hvhelper for virtual machine state unmodified (see state())
func (vf *Helper) getRawState() (define.Status, error) {
	var response rest.VMState
	endPoint := vf.Endpoint + state
	serverResponse, err := vf.get(endPoint, nil)
	if err != nil {
		if errors.Is(err, unix.ECONNREFUSED) {
			logrus.Debugf("connection refused: %s", endPoint)
		}
		return "", err
	}
	err = json.NewDecoder(serverResponse.Body).Decode(&response)
	if err != nil {
		return "", err
	}
	if err := serverResponse.Body.Close(); err != nil {
		logrus.Error(err)
	}
	return ToMachineStatus(response.State)
}

func (vf *Helper) Stop(force, wait bool) error {
	state := rest.Stop
	if force {
		state = rest.HardStop
	}
	if err := vf.stateChange(state); err != nil {
		return err
	}
	if !wait {
		return nil
	}
	waitDuration := time.Millisecond * 500
	// Wait up to 90s then hard force off
	for i := 0; i < 180; i++ {
		_, err := vf.getRawState()
		if err != nil || errors.Is(err, unix.ECONNREFUSED) {
			return nil
		}
		time.Sleep(waitDuration)
	}
	logrus.Warn("Failed to gracefully stop machine, performing hard stop")
	// we waited long enough do a hard stop
	return vf.stateChange(rest.HardStop)
}
