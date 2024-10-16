//go:build !drawin && !linux && windows

package wsl

import (
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/ports"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/utils"
	"fmt"
	gvproxy "github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

type WSLStubber struct {
	define.WSLConfig
}

func (w WSLStubber) GetDisk(userInputPath string, dirs *define.MachineDirs, imagePath *define.VMFile, vmType define.VMType, name string) error {
	// Do not do anything because we using wsl --import [rootfs.tar]
	switch {
	case userInputPath == "":
		return fmt.Errorf("Please --bootable-image [IMAGE_PATH]")
	case strings.HasPrefix(userInputPath, "http"):
		return fmt.Errorf("Do not support download image from http(s)://")
	case strings.HasPrefix(userInputPath, "docker://"):
		return fmt.Errorf("Do not support download image from docker://")
	default:
	}
	return nil
}

func (w WSLStubber) CreateVM(opts define.CreateVMOpts, mc *vmconfigs.MachineConfig) error {
	var (
		err error
	)
	callbackFuncs := machine.CleanupCallback{}
	defer callbackFuncs.CleanIfErr(&err)
	go callbackFuncs.CleanOnSignal()
	mc.WSLHypervisor = new(vmconfigs.WSLConfig)

	_ = setupWslProxyEnv()

	err = unprovisionWSL(mc, false)
	if err != nil {
		return err
	}
	dist, err := provisionWSLDist(opts, mc.ImagePath.GetPath())
	if err != nil {
		return err
	}

	unprovisionCallbackFunc := func() error {
		return unprovisionWSL(mc, true)
	}
	callbackFuncs.Add(unprovisionCallbackFunc)

	if err = configureSystem(mc, dist); err != nil {
		return err
	}

	return terminateDist(dist)
}

func configureSystem(mc *vmconfigs.MachineConfig, dist string) error {
	// Do nothing...
	return nil
}

// TODO like provisionWSL, i think this needs to be pushed to use common
// paths and types where possible
func unprovisionWSL(mc *vmconfigs.MachineConfig, ifDeleteDir bool) error {
	dist := mc.Name
	if err := terminateDist(dist); err != nil {
		logrus.Error(err)
	}
	if err := unregisterDist(dist); err != nil {
		logrus.Error(err)
	}
	if ifDeleteDir {
		vmDataDir := mc.Dirs.DataDir.GetPath()
		err := utils.GuardedRemoveAll(vmDataDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func provisionWSLDist(opts define.CreateVMOpts, distroInstallDir string) (string, error) {
	_ = os.MkdirAll(distroInstallDir, 0755)
	if err := runCmdPassThrough(FindWSL(), "--import", opts.Name, distroInstallDir, opts.UserImageFile, "--version", "2"); err != nil {
		return "", fmt.Errorf("the WSL import of guest OS failed: %w", err)
	}
	return opts.Name, nil
}

func (w WSLStubber) StopVM(mc *vmconfigs.MachineConfig, hardStop bool) error {
	if running, err := isRunning(mc.Name); !running {
		return err
	}
	dist := mc.Name

	err := wslPipe("sync", dist)
	if err != nil {
		return err
	}
	return terminateDist(dist)
}

func (w WSLStubber) MountType() vmconfigs.VolumeMountType {
	return vmconfigs.RawDisk
}

func (w WSLStubber) RequireExclusiveActive() bool {
	// WSL2 support multiple instance startup, no need exclusive active
	return false
}

func (w WSLStubber) State(mc *vmconfigs.MachineConfig) (define.Status, error) {
	running, err := isRunning(mc.Name)
	if err != nil {
		return "", err
	}
	if running {
		return define.Running, nil
	}
	return define.Stopped, nil
}

func (w WSLStubber) UpdateSSHPort(mc *vmconfigs.MachineConfig, port int) error {
	const changePort = `sed -E -i 's/^Port[[:space:]]+[0-9]+/Port %d/' /etc/ssh/sshd_config`
	dist := mc.Name
	if err := wslInvoke(dist, "sh", "-c", fmt.Sprintf(changePort, port)); err != nil {
		return fmt.Errorf("could not change SSH port for guest OS: %w", err)
	}

	return nil
}

func (w WSLStubber) UseProviderNetworkSetup() bool {
	// WSL2 do not have Use Provider NetworkSetup
	return false
}

func (w WSLStubber) StartNetworking(mc *vmconfigs.MachineConfig, cmd *gvproxy.GvproxyCommand) error {
	// WSL2 do not have StartNetworking logic
	return nil
}

func (w WSLStubber) PostStartNetworking(mc *vmconfigs.MachineConfig, noInfo bool) error {
	return nil
}

func (w WSLStubber) StartVM(mc *vmconfigs.MachineConfig) (func() error, func() error, error) {
	distName := mc.Name

	// TODO: This should be in /opt/ovmd, but for now just `echo OkImFine`
	newPort, err := ports.AllocateMachinePort()
	if err != nil {
		return nil, nil, err
	}
	success := false
	defer func() {
		if !success {
			if err := ports.ReleaseMachinePort(newPort); err != nil {
				logrus.Warnf("could not release port allocation as part of failure rollback (%d): %s", newPort, err.Error())
			}
		}
	}()

	// TODO 计算 VHDX 的大小

	str := fmt.Sprintf("/opt/ovmd", "-s", "10240", "-p", newPort)
	logrus.Infof("str %s", str)

	err = wslInvoke(distName, "echo", "OkImFine")
	if err != nil {
		err = fmt.Errorf("the WSL bootstrap script failed: %w", err)
	}

	// TODO: implement wsl2 ready Func
	readyFunc := func() error {
		return nil
	}

	releaseCmd := func() error {
		return nil
	}
	return releaseCmd, readyFunc, err
}

// TODO mount bare image into wsl
func (w WSLStubber) MountVolumesToVM(mc *vmconfigs.MachineConfig, quiet bool) error {
	return nil
}

func (w WSLStubber) Exists(distroName string) (bool, error) {
	return isWSLExist(distroName)
}

func (w WSLStubber) VMType() define.VMType {
	return define.WSLVirt
}

func isRunning(dist string) (bool, error) {
	running, err := isWSLRunning(dist)
	if err != nil {
		return false, err
	}
	return running, err
}

func (w WSLStubber) SetProviderAttrs(mc *vmconfigs.MachineConfig, opts define.SetOptions) error {
	return nil
}
