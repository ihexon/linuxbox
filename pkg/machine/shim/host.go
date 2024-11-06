package shim

import (
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/connection"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/gvproxy"
	"bauklotze/pkg/machine/lock"
	"bauklotze/pkg/machine/provider"
	"bauklotze/pkg/machine/system"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/network"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/containers/common/pkg/strongunits"
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
	"time"
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
		UserImageFile: opts.Images.BootableImage,
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
	logrus.Infof("A bootable Image provided: %s", initCmdOpts.Images.BootableImage)
	// Extract the bootable image

	// Jump into Provider's GetDisk implementation, but we can using
	// if err := diskpull.GetDisk(opts.Image, dirs, mc.ImagePath, mp.VMType(), mc.Name); err != nil {
	//		return err
	//	}
	// for simplify code, but for now keep using Provider's GetDisk implementation
	network.Reporter.SendEventToOvmJs("decompress", "running")
	if err = mp.GetDisk(initCmdOpts.Images.BootableImage, dirs, mc.ImagePath, mp.VMType(), mc.Name); err != nil {
		return err
	} else {
		network.Reporter.SendEventToOvmJs("decompress", "success")
	}
	callbackFuncs.Add(mc.ImagePath.Delete)

	if err = connection.AddSSHConnectionsToPodmanSocket(0, mc.SSH.Port, mc.SSH.IdentityPath, mc.Name, mc.SSH.RemoteUsername, opts); err != nil {
		return err
	}

	cleanup := func() error {
		machines, err := provider.GetAllMachinesAndRootfulness()
		if err != nil {
			return err
		}
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
	mc.ImageVersion = opts.ImageVersion.BootableImageVersion

	mc.DataDisk = &define.VMFile{Path: opts.Images.DataDisk}
	mc.OverlayDisk = &define.VMFile{Path: opts.Images.OverlayImage}

	mc.DataDiskVersion = opts.ImageVersion.DataDiskVersion

	network.Reporter.SendEventToOvmJs("writeConfig", "running")
	err = mc.Write()
	if err != nil {
		return err
	} else {
		network.Reporter.SendEventToOvmJs("writeConfig", "success")
		callbackFuncs.Add(mc.ConfigPath.Delete)
	}
	//err = fmt.Errorf("Test Error happend")
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

// checkExclusiveActiveVM checks if any of the machines are already running
func checkExclusiveActiveVM(provider vmconfigs.VMProvider, mc *vmconfigs.MachineConfig) error {
	// Check if any other machines are running; if so, we error
	localMachines, err := getMCsOverProviders([]vmconfigs.VMProvider{provider})
	if err != nil {
		return err
	}
	for name, localMachine := range localMachines {
		state, err := provider.State(localMachine)
		if err != nil {
			return err
		}
		if state == define.Running || state == define.Starting {
			return fmt.Errorf("unable to start %q: machine %s: %w", mc.Name, name, define.ErrVMAlreadyRunning)
		}
	}
	return nil
}

func startNetAndForwardNow(
	mc *vmconfigs.MachineConfig,
	mp vmconfigs.VMProvider,
	dirs *define.MachineDirs,
) (
	string,
	machine.APIForwardingState,
	error,
) {
	forwardSocketPath, forwardSocketState, err := startNetworking(mc, mp)
	if err != nil {
		return "", machine.NoForwarding, err
	}
	return forwardSocketPath, forwardSocketState, nil
}

func Start(mc *vmconfigs.MachineConfig, mp vmconfigs.VMProvider, dirs *define.MachineDirs, opts define.StartOptions) error {
	var err error

	callBackFuncs := machine.CleanUp()
	defer callBackFuncs.CleanIfErr(&err)
	go callBackFuncs.CleanOnSignal()

	mc.Lock()
	defer mc.Unlock()

	if err := mc.Refresh(); err != nil {
		return fmt.Errorf("reload config: %w", err)
	}

	// Don't check if provider supports parallel running machines
	// RequireExclusiveActive return false means the provider supports parallel running
	if mp.RequireExclusiveActive() {
		startLock, err := lock.GetMachineStartLock()
		if err != nil {
			return err
		}
		startLock.Lock()
		defer startLock.Unlock()

		if err := checkExclusiveActiveVM(mp, mc); err != nil {
			return err
		}
	} else {
		// still should make sure we do not start the same machine twice
		state, err := mp.State(mc)
		if err != nil {
			return err
		}

		if state == define.Running || state == define.Starting {
			return fmt.Errorf("machine %s: %w", mc.Name, define.ErrVMAlreadyRunning)
		}
	}

	// Set starting to true
	mc.Starting = true
	if err := mc.Write(); err != nil {
		logrus.Error(err)
	}
	// Set starting to false on exit
	defer func() {
		mc.Starting = false
		if err := mc.Write(); err != nil {
			logrus.Error(err)
		}
	}()

	forwardSocketPath, forwardingState, err := startNetAndForwardNow(mc, mp, dirs)
	if err != nil {
		return err
	}
	gvproxyPidFile, err := dirs.RuntimeDir.AppendToNewVMFile("gvproxy.pid", nil)
	if err != nil {
		return err
	}

	cleanGV := func() error {
		_ = system.KillProcess(machine.GlobalPIDs.GetGvproxyPID())
		logrus.Infof("--> Callback: clean gvproxy process %d", machine.GlobalPIDs.GetGvproxyPID())
		_ = gvproxyPidFile.Delete()
		_ = os.Remove(forwardSocketPath)
		return nil
	}
	callBackFuncs.Add(cleanGV)

	// Start krunkit now
	krunCmd, WaitForReady, err := mp.StartVM(mc)
	if err != nil {
		return err
	}

	cleanKRUN := func() error {
		_ = krunCmd.Process.Kill()
		logrus.Infof("--> Callback: clean krunkit process %d", krunCmd.Process.Pid)
		return nil
	}

	callBackFuncs.Add(cleanKRUN)

	if WaitForReady == nil {
		err = errors.New("no valid WaitForReady function returned")
		return err
	}

	// continue check krunkit runnning and wait ready event comming
	if err = WaitForReady(); err != nil {
		return err
	}

	err = mp.PostStartNetworking(mc, false)
	if err != nil {
		return err
	}

	//Update state
	stateF := func() (define.Status, error) {
		return mp.State(mc)
	}

	defaultBackoff := 500 * time.Millisecond
	maxBackoffs := 6

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

	// mount the volumes to the VM
	if err := mp.MountVolumesToVM(mc, false); err != nil {
		return err
	}

	err = machine.WaitAPIAndPrintInfo(
		opts.CommonOptions.ReportUrl,
		forwardSocketPath,
		forwardingState,
		mc.Name,
	)
	if err != nil {
		return err
	}

	return err
}

// Stop stops the machine
func Stop(mc *vmconfigs.MachineConfig, mp vmconfigs.VMProvider, dirs *define.MachineDirs, hardStop bool) error {
	mc.Lock()
	defer mc.Unlock()
	if err := mc.Refresh(); err != nil {
		return fmt.Errorf("reload config: %w", err)
	}

	return stopLocked(mc, mp, dirs, hardStop)
}

// stopLocked stops the machine and expects the caller to hold the machine's lock.
func stopLocked(mc *vmconfigs.MachineConfig, machineProvider vmconfigs.VMProvider, dirs *define.MachineDirs, hardStop bool) error {
	var err error
	state, err := machineProvider.State(mc)
	if err != nil {
		return err
	}
	// stopping a stopped machine is NOT an error
	if state == define.Stopped {
		return nil
	}
	if state != define.Running {
		return define.ErrWrongState
	}

	// Provider stops the machine
	if err = machineProvider.StopVM(mc, hardStop); err != nil {
		return err
	}

	// Remove Ready Socket
	readySocket, err := mc.ReadySocket()
	if err != nil {
		return err
	}
	if err := readySocket.Delete(); err != nil {
		return err
	}
	// Remove ignitionSocket Socket
	ignitionSocket, err := mc.IgnitionSocket()
	if err != nil {
		return err
	}
	if err := ignitionSocket.Delete(); err != nil {
		return err
	}

	// Stop GvProxy and remove PID file
	gvproxyPidFile, err := dirs.RuntimeDir.AppendToNewVMFile(env.Gvpid, nil)
	if err != nil {
		return err
	}
	if err := gvproxy.CleanupGVProxy(*gvproxyPidFile); err != nil {
		return fmt.Errorf("unable to clean up gvproxy: %w", err)
	}

	// Update last time up
	mc.LastUp = time.Now()
	return mc.Write()
}

// Set set configure for virtualMachine configuration
func Set(mc *vmconfigs.MachineConfig, mp vmconfigs.VMProvider, opts define.SetOptions) error {
	mc.Lock()
	defer mc.Unlock()

	if err := mc.Refresh(); err != nil {
		return fmt.Errorf("reload config: %w", err)
	}

	if opts.CPUs != 0 {
		mc.Resources.CPUs = opts.CPUs
	}

	if opts.Memory != 0 {
		mc.Resources.Memory = strongunits.MiB(opts.Memory)
	}

	if opts.Volumes != nil {
		mc.Mounts = CmdLineVolumesToMounts(opts.Volumes, mp.MountType())
	}

	if err := mp.SetProviderAttrs(mc, opts); err != nil {
		return err
	}

	// Update the configuration file last if everything earlier worked
	return mc.Write()
}
