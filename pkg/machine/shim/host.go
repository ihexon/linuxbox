//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package shim

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"time"

	"bauklotze/pkg/machine/events"
	"bauklotze/pkg/port"

	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/system"
	"bauklotze/pkg/machine/vmconfigs"

	"github.com/sirupsen/logrus"
)

// VMExists looks old machine for a machine's existence.  returns the actual config and found bool
func VMExists(name string, vmstubbers []vmconfigs.VMProvider) (*vmconfigs.MachineConfig, bool, error) {
	// Look on disk first
	mcs, err := getMCsOverProviders(vmstubbers)
	if err != nil {
		return nil, false, fmt.Errorf("failed to load machine configs: %w", err)
	}
	if mc, found := mcs[name]; found {
		return mc, true, nil
	}
	return nil, false, nil
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
		return fmt.Errorf("failed to get machine dirs: %w", err)
	}

	logrus.Infof("ConfigDir:     %s", dirs.ConfigDir.GetPath())
	logrus.Infof("DataDir:       %s", dirs.DataDir.GetPath())
	logrus.Infof("ImageCacheDir: %s", dirs.ImageCacheDir.GetPath())
	logrus.Infof("RuntimeDir:    %s", dirs.RuntimeDir.GetPath())
	logrus.Infof("LogsDir:       %s", dirs.LogsDir.GetPath())

	sshIdentityPath, err := env.GetSSHIdentityPath(define.DefaultIdentityName)
	if err != nil {
		return fmt.Errorf("failed to get ssh identity path: %w", err)
	}
	logrus.Infof("SSH identity path: %s", sshIdentityPath)

	mySSHKey, err := machine.GetSSHKeys(sshIdentityPath)
	if err != nil {
		return fmt.Errorf("failed to get ssh keys: %w", err)
	}
	logrus.Infof("SSH key: %v", mySSHKey)

	// construct a machine configure but not write into disk
	mc, err := vmconfigs.NewMachineConfig(opts, dirs, sshIdentityPath, mp.VMType())
	if err != nil {
		return fmt.Errorf("failed to create machine config: %w", err)
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
	case define.LibKrun, define.VFkit:
		imageExtension = ".raw"
	case define.WSLVirt:
		imageExtension = ""
	default:
		return fmt.Errorf("unknown VM type: %s", mp.VMType())
	}

	imagePath, err = dirs.DataDir.AppendToNewVMFile(fmt.Sprintf("%s-%s%s", opts.Name, runtime.GOARCH, imageExtension), nil)
	if err != nil {
		return fmt.Errorf("failed to append image to vm file: %w", err)
	}
	logrus.Infof("Bootable Image Path: %s", imagePath.GetPath())
	mc.ImagePath = imagePath // mc.ImagePath is the bootable copied from user provided image --boot <bootable.img.xz>

	// Generate the mc.Mounts structs from the opts.Volumes
	// mp.MountType() return vmconfigs.VirtIOFS
	mc.Mounts = CmdLineVolumesToMounts(opts.Volumes, mp.MountType())
	jsonMounts, err := json.MarshalIndent(mc.Mounts, "", "  ")
	if err != nil {
		logrus.Errorf("Failed to marshal mc.Mounts to JSON: %v", err)
	} else {
		logrus.Infof("Mounts: %s", jsonMounts)
	}

	initCmdOpts := opts
	logrus.Infof("A bootable Image provided: %s", initCmdOpts.ImagesStruct.BootableImage)

	err = mp.GetDisk(initCmdOpts.ImagesStruct.BootableImage, dirs, mc.ImagePath, mp.VMType(), mc.Name)
	if err != nil {
		return fmt.Errorf("failed to get disk: %w", err)
	}
	events.NotifyInit(events.ExtractBootImage)

	callbackFuncs.Add(func() error {
		logrus.Infof("callback: Removing image %s", mc.ImagePath.GetPath())
		return mc.ImagePath.Delete()
	})

	err = mp.CreateVM(createOpts, mc)
	if err != nil {
		return fmt.Errorf("failed to create vm: %w", err)
	}

	mc.ReportURL = &define.VMFile{Path: opts.CommonOptions.ReportURL}

	// Fill all the configure field and write into disk
	mc.ImagePath = imagePath
	mc.BootableDiskVersion = opts.ImageVerStruct.BootableImageVersion

	mc.DataDisk = &define.VMFile{Path: opts.ImagesStruct.DataDisk}

	mc.DataDiskVersion = opts.ImageVerStruct.DataDiskVersion

	err = mc.Write()
	if err != nil {
		return fmt.Errorf("failed to write machine config: %w", err)
	}
	events.NotifyInit(events.InitUpdateConfig)

	callbackFuncs.Add(func() error {
		logrus.Infof("callback: Removing Machine config %s", mc.ConfigPath.GetPath())
		return mc.ConfigPath.Delete()
	})
	return nil
}

// getMCsOverProviders loads machineconfigs from a config dir derived from the "provider".  it returns only what is known on
// disk so things like status may be incomplete or inaccurate
func getMCsOverProviders(vmstubbers []vmconfigs.VMProvider) (map[string]*vmconfigs.MachineConfig, error) {
	mcs := make(map[string]*vmconfigs.MachineConfig)
	for _, stubber := range vmstubbers {
		dirs, err := env.GetMachineDirs(stubber.VMType())
		if err != nil {
			return nil, fmt.Errorf("failed to get machine dirs: %w", err)
		}
		stubberMCs, err := vmconfigs.LoadMachinesInDir(dirs)
		if err != nil {
			return nil, fmt.Errorf("failed to load machines in dir: %w", err)
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
		return fmt.Errorf("failed to get machine state: %w", err)
	}

	if state == define.Running || state == define.Starting {
		return fmt.Errorf("machine %s: %w", mc.Name, define.ErrVMAlreadyRunning)
	}

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

	sshPort, err := port.GetFree(mc.SSH.Port)
	if err != nil {
		return fmt.Errorf("failed to get free port: %w", err)
	}

	if sshPort != mc.SSH.Port {
		if err := mc.Write(); err != nil {
			return fmt.Errorf("failed to write machine config: %w", err)
		}
		mc.SSH.Port = sshPort
	}

	gvproxyPidFile, err := dirs.RuntimeDir.AppendToNewVMFile("gvproxy.pid", nil)
	if err != nil {
		return fmt.Errorf("failed to create gvproxy pid file: %w", err)
	}

	// start gvproxy and set up the API socket forwarding
	socksInHost, forwardingState, gvcmd, err := startNetworking(mc, mp)
	// _, _, gvcmd, err := startNetworking(mc, mp)

	if err != nil {
		return fmt.Errorf("failed to start networking: %w", err)
	}

	// Start krunkit now
	logrus.Infof("--> Start krunkit....")
	krunCmd, WaitForReady, err := mp.StartVM(mc)
	if err != nil {
		return fmt.Errorf("failed to start krunkit: %w", err)
	}

	if WaitForReady == nil {
		return fmt.Errorf("no valid WaitForReady function returned")
	}

	if err = WaitForReady(); err != nil {
		return fmt.Errorf("failed to wait for ready: %w", err)
	}

	// Update state
	stateF := func() (define.Status, error) {
		return mp.State(mc)
	}

	if mp.VMType() != define.WSLVirt {
		connected, sshError, err := conductVMReadinessCheck(mc, stateF)
		if err != nil {
			return fmt.Errorf("failed to conduct vm readiness check: %w", err)
		}
		if !connected {
			msg := "machine did not transition into running state"
			if sshError != nil {
				return fmt.Errorf("%s: ssh error: %w", msg, sshError)
			}
			return errors.New(msg)
		} else {
			logrus.Infof("Machine %s SSH is ready,Using sshkey %s with %s, listen in %d", mc.Name, mc.SSH.IdentityPath, mc.SSH.RemoteUsername, mc.SSH.Port)
		}
	}

	if err = machine.WaitAPIAndPrintInfo(socksInHost, forwardingState, mc.Name); err != nil {
		return fmt.Errorf("failed to wait api and print info: %w", err)
	}

	if err = mc.UpdateLastBoot(); err != nil {
		return fmt.Errorf("failed to update last boot time: %w", err)
	}

	for {
		pids := []int32{int32(gvcmd.Process.Pid), int32(krunCmd.Process.Pid)}
		running, err := system.IsProcesSAlive(pids)
		if !running {
			_ = gvproxyPidFile.Delete()
			return fmt.Errorf("failed to start krunkit: %w", err)
		}
		time.Sleep(1 * time.Second)
	}
}
