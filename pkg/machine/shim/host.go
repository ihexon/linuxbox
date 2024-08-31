package shim

import (
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/utils"
	"errors"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"os"
	"path/filepath"
	"runtime"
)

func VMExists(name string, vmstubbers []vmconfigs.VMProvider) (*vmconfigs.MachineConfig, bool, error) {
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
func emptyfunc(p string) {

}
func Init(opts define.InitOptions, mp vmconfigs.VMProvider) error {
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
	sshIdentityPath, err := env.GetSSHIdentityPath(define.DefaultIdentityName)
	if err != nil {
		return err
	}
	sshKey, err := machine.GetSSHKeys(sshIdentityPath)
	if err != nil {
		return err
	}
	emptyfunc(sshKey)

	mc, err := vmconfigs.NewMachineConfig(opts, dirs, sshIdentityPath, mp.VMType())
	if err != nil {
		return err
	}

	mc.Version = define.MachineConfigVersion

	createOpts := define.CreateVMOpts{
		Name: opts.Name,
		Dirs: dirs,
	}

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

	dirs, err = env.GetMachineDirs(mp.VMType())
	if err != nil {
		return err
	}

	if err := mp.GetDisk(opts.Image, dirs, mc); err != nil {
		return err
	}

	// Mounts
	if mp.VMType() != define.WSLVirt {
		mc.Mounts = CmdLineVolumesToMounts(opts.Volumes, mp.MountType())
	}

	callbackFuncs.Add(mc.ImagePath.Delete)

	// Need to support ignBuilder
	// err = mp.CreateVM(createOpts, mc, &ignBuilder)
	err = mp.CreateVM(createOpts, mc)
	if err != nil {
		return err
	}

	return mc.Write()
}

func Start(mc *vmconfigs.MachineConfig, mp vmconfigs.VMProvider, dirs *define.MachineDirs, opts machine.StartOptions) error {
	return nil
}

func Reset(mps []vmconfigs.VMProvider) error {
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
		dataDirErr := utils.GuardedRemoveAll(filepath.Dir(dir.DataDir.GetPath()))
		if !errors.Is(dataDirErr, os.ErrNotExist) {
			resetErrors = multierror.Append(resetErrors, dataDirErr)
		}
		confDirErr := utils.GuardedRemoveAll(filepath.Dir(dir.ConfigDir.GetPath()))
		if !errors.Is(confDirErr, os.ErrNotExist) {
			resetErrors = multierror.Append(resetErrors, confDirErr)
		}
	}

	return resetErrors.ErrorOrNil()
}
