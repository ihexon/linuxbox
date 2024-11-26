//go:build darwin && arm64

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

// state asks vfkit for the virtual machine state. in case the vfkit
// service is not responding, we assume the service is not running
// and return a stopped status
func (vf *Helper) State() (define.Status, error) {
	vmState, err := vf.getRawState()
	if err == nil {
		return vmState, nil
	}
	if errors.Is(err, unix.ECONNREFUSED) || errors.Is(err, unix.ECONNRESET) {
		return define.Stopped, nil
	}
	return "", err
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
