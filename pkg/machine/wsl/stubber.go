package wsl

import (
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/shim/diskpull"
)

type WSLStubber struct {
	define.WSLConfig
}

func (w WSLStubber) Exists(name string) (bool, error) {
	return isWSLExist(env.WithBugBoxPrefix(name))
}

func (w WSLStubber) VMType() define.VMType {
	return define.WSLVirt
}

// 优先选择 userInputPath 指向的 rootfs
func (w WSLStubber) GetDisk(userInputPath string, mc *define.MachineConfig) error {
	if userInputPath != "" {
		// userInputPath 是用户指定的 rootfs 路径
		// mc.ImagePath 实际上是 rootfs 的路径
		return diskpull.GetDisk(userInputPath, mc.Dirs, mc.ImagePath, w.VMType(), mc.Name)
	}
	return nil
}

// TODO: checkAndInstallWSL
func (w WSLStubber) CreateVM(opts define.CreateVMOpts, mc *define.MachineConfig) error {
	checkAndInstallWSL(opts.ReExec)
	return nil
}

func (w WSLStubber) StopVM(mc *define.MachineConfig, hardStop bool) error {
	dist := env.WithBugBoxPrefix(mc.Name)
	return terminateDist(dist)
}

func isRunning(name string) (bool, error) {
	dist := env.WithBugBoxPrefix(name)
	running, err := isWSLRunning(dist)
	if err != nil {
		return false, err
	}

	// TODO: isPodmanAPI Running
	//sysd := false
	//if wsl {
	//	sysd, err = isSystemdRunning(dist)
	//
	//	if err != nil {
	//		return false, err
	//	}
	// }

	return running, err
}

func (w WSLStubber) Remove(mc *define.MachineConfig) ([]string, func() error, error) {

	wslRemoveFunc := func() error {
		if err := runCmdPassThrough(FindWSL(), "--unregister", env.WithBugBoxPrefix(mc.Name)); err != nil {
			return err
		}
		return nil
	}

	return []string{}, wslRemoveFunc, nil
}
