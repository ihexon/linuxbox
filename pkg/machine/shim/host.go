package shim

import (
	fileutils "bauklotze/pkg/ioutils"
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/vmconfig"
	"errors"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"os"
	"path/filepath"
	"runtime"
)

func VMExists(name string, vmstubbers []define.VMProvider) (*define.MachineConfig, bool, error) {
	// Check with the provider hypervisor
	for _, vmstubber := range vmstubbers {
		exists, err := vmstubber.Exists(name)
		if err != nil {
			return nil, false, err
		}
		if exists {
			return nil, true, fmt.Errorf("vm %q already exists on hypervisor", name)
		}
	}
	return nil, false, nil
}

func Init(opts define.InitOptions, mp define.VMProvider) error {
	var (
		imageExtension string
		err            error
		imagePath      *define.VMFile
	)
	// Empty callbackFuncs arraylist
	callbackFuncs := machine.CleanupFuncs()
	defer callbackFuncs.CleanIfErr(&err)
	go callbackFuncs.CleanOnSignal()

	dirs, err := env.GetMachineDirs(mp.VMType())
	if err != nil {
		return err
	}

	mc, err := vmconfig.NewMachineConfig(opts, dirs, mp.VMType())
	if err != nil {
		return err
	}

	mc.Version = define.MachineConfigVersion

	switch mp.VMType() {
	case define.QemuVirt:
		imageExtension = ".qcow2"
	case define.AppleHvVirt, define.LibKrun:
		imageExtension = ".raw"
	case define.HyperVVirt:
		imageExtension = ".vhdx"
	case define.WSLVirt:
		imageExtension = ""
	default:
		return fmt.Errorf("unknown VM type: %s", mp.VMType())
	}

	imagePath, err = dirs.DataDir.AppendToNewVMFile(fmt.Sprintf("%s-%s%s", opts.Name, runtime.GOARCH, imageExtension))
	mc.ImagePath = imagePath

	// TODO: 实现 GetDisk， 下载 rootfs
	if err := mp.GetDisk(opts.Image, mc); err != nil {
		return err
	}

	callbackFuncs.Add(mc.ImagePath.Delete)

	// TODO: CreateVM
	//err = mp.CreateVM(createOpts, mc, &ignBuilder)
	//if err != nil {
	//	return err
	//}

	return mc.Write()
}

func Reset(mps []define.VMProvider) error {
	var resetErrors *multierror.Error
	// 注意 define 是配置模板，不存储数据
	var removeDirs []*define.MachineDirs

	for _, mp := range mps {
		// env.GetMachineDirs return .local .config ~
		d, err := env.GetMachineDirs(mp.VMType())
		if err != nil {
			resetErrors = multierror.Append(resetErrors, err)
			continue
		}
		if err != nil {
			resetErrors = multierror.Append(resetErrors, err)
			continue
		}
		removeDirs = append(removeDirs, d)
	}

	for _, dir := range removeDirs {
		dataDirErr := fileutils.GuardedRemoveAll(filepath.Dir(dir.DataDir.GetPath()))
		if !errors.Is(dataDirErr, os.ErrNotExist) {
			resetErrors = multierror.Append(resetErrors, dataDirErr)
		}
		confDirErr := fileutils.GuardedRemoveAll(filepath.Dir(dir.ConfigDir.GetPath()))
		if !errors.Is(confDirErr, os.ErrNotExist) {
			resetErrors = multierror.Append(resetErrors, confDirErr)
		}
	}

	return resetErrors.ErrorOrNil()
}
