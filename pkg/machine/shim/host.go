//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package shim

import (
	"context"
	"fmt"
	"runtime"
	"time"

	allFlag "bauklotze/pkg/machine/allflag"
	"bauklotze/pkg/machine/callback"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/events"
	"bauklotze/pkg/machine/helper"
	"bauklotze/pkg/machine/io"
	sshService "bauklotze/pkg/machine/ssh/service"
	"bauklotze/pkg/machine/vmconfig"
	"bauklotze/pkg/machine/volumes"
	"bauklotze/pkg/port"
	"bauklotze/pkg/ssh"
	"bauklotze/pkg/system"

	"github.com/prashantgupta24/mac-sleep-notifier/notifier"
	"github.com/sirupsen/logrus"
)

func GetVMConf(name string, stubs []vmconfig.VMProvider) (*vmconfig.MachineConfig, error) {
	// Look on disk first
	mcs, err := getMCsOverProviders(stubs)
	if err != nil {
		return nil, fmt.Errorf("failed to load machine configs: %w", err)
	}
	if mc, found := mcs[name]; found {
		return mc, nil
	}
	return nil, fmt.Errorf("machine %s not found", name)
}

func Init(mp vmconfig.VMProvider) error {
	var err error
	callbackFuncs := callback.CleanUp()
	// Clean file when machine init occurs error
	defer callbackFuncs.CleanIfErr(&err)
	// Clean file when machine init catch signal
	go callbackFuncs.CleanOnSignal()

	dirs, err := vmconfig.GetMachineDirs(mp.VMType())
	if err != nil {
		return fmt.Errorf("failed to get machine dirs: %w", err)
	}
	logrus.Infof("ConfigDir:     %s", dirs.ConfigDir.GetPath())
	logrus.Infof("DataDir:       %s", dirs.DataDir.GetPath())
	logrus.Infof("TmpDir:    	 %s", dirs.TmpDir.GetPath())
	logrus.Infof("LogsDir:       %s", dirs.LogsDir.GetPath())

	sshKey, err := vmconfig.GetSSHIdentityPath(mp.VMType())
	if err != nil {
		return fmt.Errorf("failed to get ssh identity path: %w", err)
	}

	logrus.Infof("Generate ssh keys and write into %s", sshKey.GetPath())
	_, err = ssh.GetSSHKeys(sshKey)
	if err != nil {
		return fmt.Errorf("failed to get ssh keys: %w", err)
	}

	mc, err := vmconfig.NewMachineConfig(dirs, sshKey, mp.VMType())
	if err != nil {
		return fmt.Errorf("failed to create machine config: %w", err)
	}

	// Bootable Images Logic
	bootableImage, err := dirs.DataDir.AppendToNewVMFile(fmt.Sprintf("%s-%s.%s", allFlag.VMName, runtime.GOARCH, "raw"))
	if err != nil {
		return fmt.Errorf("failed to append image to vm file: %w", err)
	}
	logrus.Infof("Bootable Image path: %s", bootableImage.GetPath())
	mc.Bootable.Image = bootableImage
	mc.Bootable.Version = allFlag.BootableImageVersion
	if err = mp.ExtractBootable(allFlag.BootableImage, mc); err != nil {
		return fmt.Errorf("failed to get disk: %w", err)
	}
	events.NotifyInit(events.ExtractBootImage)
	callbackFuncs.Add(func() error {
		logrus.Warnf("Clean bootable image %s du to error: %v", mc.Bootable.Image.GetPath(), err)
		return mc.Bootable.Image.Delete(true)
	})
	if err = mp.CreateVMConfig(mc); err != nil {
		return fmt.Errorf("failed to create vm: %w", err)
	}

	// DataDisk Images Logic
	dataImage, err := dirs.DataDir.AppendToNewVMFile(fmt.Sprintf("%s-%s-%s.%s", allFlag.VMName, runtime.GOARCH, "data", "raw"))
	if err != nil {
		return fmt.Errorf("failed to append image to vm file: %w", err)
	}
	logrus.Infof("Data Image path: %s", dataImage.GetPath())
	mc.DataDisk.Image = dataImage
	mc.DataDisk.Version = allFlag.DataDiskVersion
	if err = helper.CreateAndResizeDisk(mc.DataDisk.Image, define.DefaultDataImageSizeGB); err != nil {
		return fmt.Errorf("failed to create and resize disk: %w", err)
	}

	// Volumes Logic
	mc.Mounts = volumes.CmdLineVolumesToMounts(allFlag.Volumes)

	// set ReportURL into machine configure
	if allFlag.ReportURL != "" {
		mc.ReportURL = &io.VMFile{Path: allFlag.ReportURL}
	}

	// Write the machine configure as json into mc.ConfigPath.GetPath()
	logrus.Infof("Write machine configure to %s", mc.ConfigPath.GetPath())
	mc.Created = time.Now()
	if err = mc.Write(); err != nil {
		return fmt.Errorf("write machine configure %s failed", mc.ConfigPath)
	}
	callbackFuncs.Add(func() error {
		logrus.Warnf("Clean machine configure %s du to error: %v", mc.ConfigPath.GetPath(), err)
		return mc.ConfigPath.Delete(true)
	})

	return nil
}

// getMCsOverProviders loads machineconfigs from a config dir derived from the "provider".
// it returns only what is known on disk so things like status may be incomplete or inaccurate
func getMCsOverProviders(stubs []vmconfig.VMProvider) (map[string]*vmconfig.MachineConfig, error) {
	mcs := make(map[string]*vmconfig.MachineConfig)
	for _, stb := range stubs {
		dirs, err := vmconfig.GetMachineDirs(stb.VMType())
		if err != nil {
			return nil, fmt.Errorf("failed to get machine dirs: %w", err)
		}
		stbMcs, err := vmconfig.LoadMachinesInDir(dirs)
		if err != nil {
			return nil, fmt.Errorf("failed to load machines in dir: %w", err)
		}
		for mcName, mc := range stbMcs {
			if _, ok := mcs[mcName]; !ok {
				mcs[mcName] = mc
			}
		}
	}
	return mcs, nil
}

func startNetworkProvider(mc *vmconfig.MachineConfig) error {
	// p is the port maybe changed different from the default port which 6123
	p, err := port.GetFree(mc.SSH.Port)
	if err != nil {
		return fmt.Errorf("failed to get free port: %w", err)
	}
	// set the port to the machine configure
	mc.SSH.Port = p
	podmanAPISocksInHost, podmanAPISocksInGuest, err := startNetworking(mc)
	if err != nil {
		return fmt.Errorf("failed to start network provider: %w", err)
	}
	logrus.Infof("Forwarding podman api socket Host:%q to Guest:%q", podmanAPISocksInHost.GetPath(), podmanAPISocksInGuest.GetPath())

	return nil
}

func tryKillHyperVisorBeforeRun(mc *vmconfig.MachineConfig) {
	f := mc.Dirs.Hypervisor.Bin.GetPath()
	proc, _ := system.FindProcessByPath(f)
	if proc != nil {
		logrus.Warnf("Find running %s process, this should never happen, try to kill", f)
		_ = proc.Kill()
	}
}

func startVMProvider(mc *vmconfig.MachineConfig) error {
	tryKillHyperVisorBeforeRun(mc)
	provider := mc.VMProvider
	logrus.Infof("Start VM provider: %s", provider.VMType())
	return provider.StartVM(mc) //nolint:wrapcheck
}

// Start the machine It will start network provider and hypervisor:
//
// 1. start network provider which is gvproxy and save the exec.cmd into mc struct
//
// 2. start hypervisor which is krunkit(arm64)/vfkit(x86_64) save the exec.cmd into mc struct
//
// Note: this function is a non-block function, it will return immediately after start the network provider and hypervisor
func Start(ctx context.Context, mc *vmconfig.MachineConfig) (context.Context, error) {
	// First start network provider which provided by gvproxy
	err := startNetworkProvider(mc)
	if err != nil {
		return nil, fmt.Errorf("failed to start network provider: %w", err)
	}

	// Start HyperVisor which provided by krunkit(arm64)/vfkit(x86_64)
	err = startVMProvider(mc)
	if err != nil {
		return nil, fmt.Errorf("failed to start vm provider: %w", err)
	}

	err = mc.Write()
	if err != nil {
		return nil, fmt.Errorf("failed to write machine config: %w", err)
	}

	return ctx, nil
}

// SleepNotifier start Sleep Notifier and dispatch tasks
func SleepNotifier(mc *vmconfig.MachineConfig) {
	notifierCh := notifier.GetInstance().Start()
	for { //nolint:gosimple
		select {
		case activity := <-notifierCh:
			if activity.Type == notifier.Awake {
				logrus.Infof("machine awake, dispatch tasks")
				if err := TimeSync(mc); err != nil {
					logrus.Errorf("Failed to sync timestamp: %v", err)
				}
			}
		}
	}
}

func TimeSync(mc *vmconfig.MachineConfig) error {
	return sshService.DoTimeSync(mc) //nolint:wrapcheck
}

func DiskSync(mc *vmconfig.MachineConfig) error {
	return sshService.DoSync(mc) //nolint:wrapcheck
}
