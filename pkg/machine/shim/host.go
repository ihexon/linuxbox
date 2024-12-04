//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package shim

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"time"

	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/connection"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/provider"
	"bauklotze/pkg/machine/system"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/network"

	"github.com/sirupsen/logrus"
)

// VMExists looks old machine for a machine's existence.  returns the actual config and found bool
func VMExists(name string, vmstubbers []vmconfigs.VMProvider) (*vmconfigs.MachineConfig, bool, error) {
	// Look on disk first
	mcs, err := getMCsOverProviders(vmstubbers)
	if err != nil {
		return nil, false, err
	}
	if mc, found := mcs[name]; found {
		return mc, true, nil
	}
	return nil, false, err
}

func Init(opts define.InitOptions, mp vmconfigs.VMProvider) error {
	var (
		imageExtension string
		err            error
		imagePath      *define.VMFile
	)

	callbackFuncs := machine.CleanUp()
	defer callbackFuncs.CleanIfErr(&err)
	go callbackFuncs.CleanOnSignal()

	dirs, err := env.GetMachineDirs(mp.VMType())
	if err != nil {
		return err
	}

	//	dirs := define.MachineDirs{
	//		ConfigDir:     configDirFile, // ${BauklotzeHomePath}/config/{wsl,libkrun,qemu,hyper...}
	//		DataDir:       dataDirFile,   // ${BauklotzeHomePath}/data/{wsl2,libkrun,qemu,hyper...}
	//		ImageCacheDir: imageCacheDir, // ${BauklotzeHomePath}/data/{wsl2,libkrun,qemu,hyper...}/cache
	//		RuntimeDir:    rtDirFile,     // ${BauklotzeHomePath}/tmp/
	//		LogsDir:       logsDirVMFile, // ${BauklotzeHomePath}/logs
	//	}
	logrus.Infof("ConfigDir:     %s", dirs.ConfigDir.GetPath())
	logrus.Infof("DataDir:       %s", dirs.DataDir.GetPath())
	logrus.Infof("ImageCacheDir: %s", dirs.ImageCacheDir.GetPath())
	logrus.Infof("RuntimeDir:    %s", dirs.RuntimeDir.GetPath())
	logrus.Infof("LogsDir:       %s", dirs.LogsDir.GetPath())

	sshIdentityPath, err := env.GetSSHIdentityPath(define.DefaultIdentityName)
	if err != nil {
		return err
	}
	logrus.Infof("SSH identity path: %s", sshIdentityPath)

	mySSHKey, err := machine.GetSSHKeys(sshIdentityPath)
	if err != nil {
		return err
	}
	logrus.Infof("SSH key: %v", mySSHKey)

	// construct a machine configure but not write into disk
	mc, err := vmconfigs.NewMachineConfig(opts, dirs, sshIdentityPath, mp.VMType())
	if err != nil {
		return err
	}

	// machine configure json,version always be as 1
	mc.Version = define.MachineConfigVersion

	createOpts := define.CreateVMOpts{
		// Distro Name : machine init [distro_name]
		Name: opts.Name,
		Dirs: dirs,
		// UserImageFile: Image Path form machine init --image [rootfs.tar]
		UserImageFile: opts.ImagesStruct.BootableImage,
	}

	switch mp.VMType() {
	case define.LibKrun:
		imageExtension = ".raw"
	case define.WSLVirt:
		imageExtension = ""
	default:
		return fmt.Errorf("unknown VM type: %s", mp.VMType())
	}

	imagePath, err = dirs.DataDir.AppendToNewVMFile(fmt.Sprintf("%s-%s%s", opts.Name, runtime.GOARCH, imageExtension), nil)
	if err != nil {
		return err
	}
	logrus.Infof("Bootable Image Path: %s", imagePath.GetPath())
	mc.ImagePath = imagePath // mc.ImagePath is the bootable copied from user provided image --boot <bootable.img.xz>

	// Generate the mc.Mounts structs from the opts.Volumes
	mc.Mounts = CmdLineVolumesToMounts(opts.Volumes, mp.MountType())
	jsonMounts, err := json.MarshalIndent(mc.Mounts, "", "  ")
	if err != nil {
		logrus.Errorf("Failed to marshal mc.Mounts to JSON: %v", err)
	} else {
		logrus.Infof("Mounts: %s", jsonMounts)
	}

	initCmdOpts := opts
	logrus.Infof("A bootable Image provided: %s", initCmdOpts.ImagesStruct.BootableImage)
	// Extract the bootable image

	// Jump into Provider's GetDisk implementation, but we can using
	// if err := diskpull.GetDisk(opts.Image, dirs, mc.ImagePath, mp.VMType(), mc.Name); err != nil {
	//		return err
	//	}
	// for simplify code, but for now keep using Provider's GetDisk implementation
	network.Reporter.SendEventToOvmJs("decompress", "running")
	if err = mp.GetDisk(initCmdOpts.ImagesStruct.BootableImage, dirs, mc.ImagePath, mp.VMType(), mc.Name); err != nil {
		return err
	} else {
		network.Reporter.SendEventToOvmJs("decompress", "success")
	}

	callbackFuncs.Add(func() error {
		logrus.Infof("--> Callback: Removing image %s", mc.ImagePath.GetPath())
		return mc.ImagePath.Delete()
	})

	if err = connection.AddSSHConnectionsToPodmanSocket(0, mc.SSH.Port, mc.SSH.IdentityPath, mc.Name, mc.SSH.RemoteUsername, opts); err != nil {
		return err
	}

	cleanup := func() error {
		machines, err := provider.GetAllMachinesAndRootfulness()
		if err != nil {
			return err
		}
		logrus.Infof("--> Callback: Removing connections for %s", mc.Name)
		return connection.RemoveConnections(machines, mc.Name+"-root")
	}
	callbackFuncs.Add(cleanup)

	err = mp.CreateVM(createOpts, mc)
	if err != nil {
		return err
	}

	mc.ReportURL = &define.VMFile{Path: opts.CommonOptions.ReportUrl}

	// Fill all the configure field and write into disk
	mc.ImagePath = imagePath
	mc.BootableDiskVersion = opts.ImageVerStruct.BootableImageVersion

	mc.DataDisk = &define.VMFile{Path: opts.ImagesStruct.DataDisk}

	mc.DataDiskVersion = opts.ImageVerStruct.DataDiskVersion

	network.Reporter.SendEventToOvmJs("writeConfig", "running")
	err = mc.Write()
	if err != nil {
		return err
	} else {
		network.Reporter.SendEventToOvmJs("writeConfig", "success")
		// callbackFuncs.Add(mc.ConfigPath.Delete)
		callbackFuncs.Add(func() error {
			logrus.Infof("--> Callback: Removing Machine config %s", mc.ConfigPath.GetPath())
			return mc.ConfigPath.Delete()
		})
	}
	// err = fmt.Errorf("Test Error happend")
	return err
}

// getMCsOverProviders loads machineconfigs from a config dir derived from the "provider".  it returns only what is known on
// disk so things like status may be incomplete or inaccurate
func getMCsOverProviders(vmstubbers []vmconfigs.VMProvider) (map[string]*vmconfigs.MachineConfig, error) {
	mcs := make(map[string]*vmconfigs.MachineConfig)
	for _, stubber := range vmstubbers {
		dirs, err := env.GetMachineDirs(stubber.VMType())
		if err != nil {
			return nil, err
		}
		stubberMCs, err := vmconfigs.LoadMachinesInDir(dirs)
		if err != nil {
			return nil, err
		}
		for mcName, mc := range stubberMCs {
			if _, ok := mcs[mcName]; !ok {
				mcs[mcName] = mc
			}
		}
	}
	return mcs, nil
}

func Start(ctx context.Context, mc *vmconfigs.MachineConfig, mp vmconfigs.VMProvider, dirs *define.MachineDirs, opts define.StartOptions) error {
	var err error

	if err := mc.Refresh(); err != nil {
		return fmt.Errorf("reload config: %w", err)
	}

	state, err := mp.State(mc)
	if err != nil {
		return err
	}

	if state == define.Running || state == define.Starting {
		return fmt.Errorf("machine %s: %w", mc.Name, define.ErrVMAlreadyRunning)
	}

	logrus.Infof("Require machine lock, if there is any other operation on this machine, it will be blocked")
	mc.Lock()
	logrus.Infof("Machine lock require success")

	defer mc.Unlock()

	// Set starting to true
	mc.Starting = true
	if err = mc.Write(); err != nil {
		logrus.Error(err)
	}

	// Set starting to false on exit
	defer func() {
		mc.Starting = false
		if err = mc.Write(); err != nil {
			logrus.Error(err)
		}
	}()

	gvproxyPidFile, err := dirs.RuntimeDir.AppendToNewVMFile(fmt.Sprintf("gvproxy.pid", mc.Name), nil)
	if err != nil {
		return err
	}

	// start gvproxy and set up the API socket forwarding
	socksInHost, forwardingState, gvcmd, err := startNetworking(mc, mp)
	if err != nil {
		return err
	}

	// Start krunkit now
	logrus.Infof("Start krunkit....")
	krunCmd, WaitForReady, err := mp.StartVM(mc)
	if err != nil {
		return err
	}

	if WaitForReady == nil {
		err = errors.New("no valid WaitForReady function returned")
		return err
	}

	if err = WaitForReady(); err != nil {
		return err
	}

	// Update state
	stateF := func() (define.Status, error) {
		return mp.State(mc)
	}
	//
	defaultBackoff := 500 * time.Millisecond
	maxBackoffs := 3

	if mp.VMType() != define.WSLVirt {
		connected, sshError, err := conductVMReadinessCheck(mc, maxBackoffs, defaultBackoff, stateF)
		if err != nil {
			return err
		}
		if !connected {
			msg := "machine did not transition into running state"
			if sshError != nil {
				return fmt.Errorf("%s: ssh error: %v", msg, sshError)
			}
			return errors.New(msg)
		} else {
			logrus.Infof("Machine %s SSH is ready,Using sshkey %s with %s, listen in %d", mc.Name, mc.SSH.IdentityPath, mc.SSH.RemoteUsername, mc.SSH.Port)
		}
	}

	if err = machine.WaitAPIAndPrintInfo(socksInHost, forwardingState, mc.Name); err != nil {
		return err
	}

	running := false
	for {
		pids := []int32{int32(gvcmd.Process.Pid), int32(krunCmd.Process.Pid)}
		running, err = system.IsProcesSAlive(pids)
		if !running {
			_ = gvproxyPidFile.Delete()
			return fmt.Errorf("%v", err)
		}
	}
}
