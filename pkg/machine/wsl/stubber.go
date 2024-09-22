//go:build !drawin && !linux && windows

package wsl

import (
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/machineDefine"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/utils"
	"fmt"
	gvproxy "github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/sirupsen/logrus"
	"strings"
)

type WSLStubber struct {
	machineDefine.WSLConfig
}

func (w WSLStubber) GetDisk(userInputPath string,
	dirs *machineDefine.MachineDirs,
	vmType machineDefine.VMType,
	mc *vmconfigs.MachineConfig) error {
	// Do not do anything because we using wsl --import [rootfs.tar]
	switch {
	case userInputPath == "":
		return fmt.Errorf("Please --image [IMAGE_PATH]")
	case strings.HasPrefix(userInputPath, "http"):
		return fmt.Errorf("Do not support download image from http(s)://")
	case strings.HasPrefix(userInputPath, "docker://"):
		return fmt.Errorf("Do not support download image from docker://")
	default:
	}
	return nil
}

func (w WSLStubber) CreateVM(opts machineDefine.CreateVMOpts, mc *vmconfigs.MachineConfig) error {
	var (
		err error
	)
	callbackFuncs := machine.CleanupCallback{}
	defer callbackFuncs.CleanIfErr(&err)
	go callbackFuncs.CleanOnSignal()
	mc.WSLHypervisor = new(vmconfigs.WSLConfig)

	const prompt = "Importing operating system into WSL (this may take a few minutes on a new WSL install)..."

	_ = setupWslProxyEnv()
	dist, err := provisionWSLDist(opts, mc.ImagePath.GetPath(), prompt)
	if err != nil {
		return err
	}

	unprovisionCallbackFunc := func() error {
		return unprovisionWSL(mc)
	}
	callbackFuncs.Add(unprovisionCallbackFunc)
	logrus.Infof("Configuring system...")
	return terminateDist(dist)
}

// TODO like provisionWSL, i think this needs to be pushed to use common
// paths and types where possible
func unprovisionWSL(mc *vmconfigs.MachineConfig) error {
	dist := mc.Name
	if err := terminateDist(dist); err != nil {
		logrus.Error(err)
	}
	if err := unregisterDist(dist); err != nil {
		logrus.Error(err)
	}
	vmDataDir := mc.Dirs.DataDir.GetPath()
	return utils.GuardedRemoveAll(vmDataDir)
}

func provisionWSLDist(opts machineDefine.CreateVMOpts, distroInstallDir string, prompt string) (string, error) {
	if err := runCmdPassThrough(FindWSL(), "--import", opts.Name, distroInstallDir, opts.UserImageFile, "--version", "2"); err != nil {
		return "", fmt.Errorf("the WSL import of guest OS failed: %w", err)
	}
	return opts.Name, nil

}

func (w WSLStubber) StopVM(mc *vmconfigs.MachineConfig, hardStop bool) error {
	if running, err := isRunning(mc.Name); !running {
		return err
	}
	dist := (mc.Name)

	// Stop user-mode networking if enabled
	//if err := stopUserModeNetworking(mc); err != nil {
	//	fmt.Fprintf(os.Stderr, "Could not cleanly stop user-mode networking: %s\n", err.Error())
	//}

	//if err := machine.StopWinProxy(mc.Name, vmtype); err != nil {
	//	fmt.Fprintf(os.Stderr, "Could not stop API forwarding service (win-sshproxy.exe): %s\n", err.Error())
	//}

	err := wslPipe("sync", dist)
	if err != nil {
		return err
	}
	return terminateDist(dist)
}

func (w WSLStubber) MountType() vmconfigs.VolumeMountType {
	//TODO implement me
	panic("implement me")
}

func (w WSLStubber) RequireExclusiveActive() bool {
	//TODO implement me
	return false
}

func (w WSLStubber) State(mc *vmconfigs.MachineConfig, bypass bool) (machineDefine.Status, error) {
	running, err := isRunning(mc.Name)
	if err != nil {
		return "", err
	}
	if running {
		return machineDefine.Running, nil
	}
	return machineDefine.Stopped, nil
}

func (w WSLStubber) UpdateSSHPort(mc *vmconfigs.MachineConfig, port int) error {
	//TODO implement me
	panic("implement me")
}

func (w WSLStubber) UseProviderNetworkSetup() bool {
	return false
}

func (w WSLStubber) StartNetworking(mc *vmconfigs.MachineConfig, cmd *gvproxy.GvproxyCommand) error {
	//TODO implement me
	return nil
}

func (w WSLStubber) PostStartNetworking(mc *vmconfigs.MachineConfig, noInfo bool) error {
	return nil
}

func (w WSLStubber) StartVM(mc *vmconfigs.MachineConfig) (func() error, func() error, error) {
	dist := (mc.Name)

	err := wslInvoke(dist, "echo", "OkImFine")
	if err != nil {
		err = fmt.Errorf("the WSL bootstrap script failed: %w", err)
	}

	readyFunc := func() error {
		return nil
	}
	return nil, readyFunc, err
}

// TODO mount bare image into wsl
func (w WSLStubber) MountVolumesToVM(mc *vmconfigs.MachineConfig, quiet bool) error {

	return nil
}

func (w WSLStubber) Exists(distroName string) (bool, error) {
	return isWSLExist(distroName)
}

func (w WSLStubber) VMType() machineDefine.VMType {
	return machineDefine.WSLVirt
}

func (w WSLStubber) Remove(mc *vmconfigs.MachineConfig) ([]string, func() error, error) {
	wslRemoveFunc := func() error {
		if err := runCmdPassThrough(FindWSL(), "--unregister", mc.Name); err != nil {
			return err
		}
		return nil
	}

	return []string{}, wslRemoveFunc, nil
}

func isRunning(dist string) (bool, error) {
	running, err := isWSLRunning(dist)
	if err != nil {
		return false, err
	}
	return running, err
}
