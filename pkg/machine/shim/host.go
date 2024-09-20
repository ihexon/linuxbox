package shim

import (
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/connection"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/gvproxy"
	"bauklotze/pkg/machine/lock"
	"bauklotze/pkg/machine/machineDefine"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/utils"
	"errors"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
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

// VMExistsUsingProvider looks across given providers for a machine's existence. returns the actual config and found bool
func VMExistsUsingProvider(name string, vmstubbers []vmconfigs.VMProvider) (*vmconfigs.MachineConfig, bool, error) {
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
	// print ssh keys
}

func Init(opts machineDefine.InitOptions, mp vmconfigs.VMProvider) error {
	var (
		imageExtension string
		err            error
		imagePath      *machineDefine.VMFile
	)

	// Empty callbackFuncs arraylist
	callbackFuncs := machine.CleanupFuncs()
	defer callbackFuncs.CleanIfErr(&err)
	go callbackFuncs.CleanOnSignal()

	dirs, err := env.GetMachineDirs(mp.VMType())
	if err != nil {
		return err
	}
	sshIdentityPath, err := env.GetSSHIdentityPath(machineDefine.DefaultIdentityName)
	if err != nil {
		return err
	}
	sshKey, err := machine.GetSSHKeys(sshIdentityPath)
	if err != nil {
		return err
	}

	// TODO: write sshkey to rootfs
	emptyfunc(sshKey)

	// construct a machine configure but not write into disk
	mc, err := vmconfigs.NewMachineConfig(opts, dirs, sshIdentityPath, mp.VMType())
	if err != nil {
		return err
	}
	// machine configure json,version always be as 1
	mc.Version = machineDefine.MachineConfigVersion

	createOpts := machineDefine.CreateVMOpts{
		Name: opts.Name,
		Dirs: dirs,
	}

	switch mp.VMType() {
	case machineDefine.LibKrun:
		imageExtension = ".raw"
	case machineDefine.WSLVirt:
		imageExtension = ""
	default:
		return fmt.Errorf("unknown VM type: %s", mp.VMType())
	}

	imagePath, err = dirs.DataDir.AppendToNewVMFile(fmt.Sprintf("%s-%s%s", opts.Name, runtime.GOARCH, imageExtension))
	mc.ImagePath = imagePath

	// Mounts
	if mp.VMType() != machineDefine.WSLVirt {
		mc.Mounts = CmdLineVolumesToMounts(opts.Volumes, mp.MountType())
	}

	{
		if err := mp.GetDisk(opts.Image, dirs, mc); err != nil {
			return err
		}
		logrus.Infof("--> imagePath is %q", imagePath.GetPath())
		callbackFuncs.Add(mc.ImagePath.Delete)

	}

	{
		// TODO AddSSHConnectionToPodmanSocket could take an machineconfig instead
		if err := connection.AddSSHConnectionsToPodmanSocket(mc.HostUser.UID, mc.SSH.Port, mc.SSH.IdentityPath, mc.Name, mc.SSH.RemoteUsername, opts); err != nil {
			return err
		}
		cleanup := func() error {
			return connection.RemoveConnections(mc.Name, mc.Name+"-root")
		}
		callbackFuncs.Add(cleanup)
	}

	err = mp.CreateVM(createOpts, mc)
	if err != nil {
		return err
	}

	mc.EvtSockPath = &machineDefine.VMFile{Path: opts.SendEvt}
	mc.TwinPid = opts.TwinPid
	mc.ImageVersion = opts.ImageVersion

	return mc.Write()
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

func GetMCsOverProviders(vmstubbers []vmconfigs.VMProvider) (map[string]*vmconfigs.MachineConfig, error) {
	return getMCsOverProviders(vmstubbers)
}

// checkExclusiveActiveVM checks if any of the machines are already running
func checkExclusiveActiveVM(provider vmconfigs.VMProvider, mc *vmconfigs.MachineConfig) error {
	// Check if any other machines are running; if so, we error
	localMachines, err := getMCsOverProviders([]vmconfigs.VMProvider{provider})
	if err != nil {
		return err
	}
	for name, localMachine := range localMachines {
		state, err := provider.State(localMachine, false)
		if err != nil {
			return err
		}
		if state == machineDefine.Running || state == machineDefine.Starting {
			return fmt.Errorf("unable to start %q: machine %s: %w", mc.Name, name, machineDefine.ErrVMAlreadyRunning)
		}
	}
	return nil
}

func Start(mc *vmconfigs.MachineConfig, mp vmconfigs.VMProvider, dirs *machineDefine.MachineDirs, opts machineDefine.StartOptions) error {
	defaultBackoff := 500 * time.Millisecond
	maxBackoffs := 6
	noInfo := opts.NoInfo
	mc.Lock()
	defer mc.Unlock()
	if err := mc.Refresh(); err != nil {
		return fmt.Errorf("reload config: %w", err)
	}

	// Don't check if provider supports parallel running machines
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
		state, err := mp.State(mc, false)
		if err != nil {
			return err
		}

		if state == machineDefine.Running || state == machineDefine.Starting {
			return fmt.Errorf("machine %s: %w", mc.Name, machineDefine.ErrVMAlreadyRunning)
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

	gvproxyPidFile, err := dirs.RuntimeDir.AppendToNewVMFile("gvproxy.pid")
	if err != nil {
		return err
	}

	// start gvproxy and set up the API socket forwarding
	forwardSocketPath, forwardingState, err := startNetworking(mc, mp)
	if err != nil {
		return err
	}

	callBackFuncs := machine.CleanupFuncs()
	defer callBackFuncs.CleanIfErr(&err)
	go callBackFuncs.CleanOnSignal()

	// Clean up gvproxy if start fails
	cleanGV := func() error {
		return gvproxy.CleanupGVProxy(*gvproxyPidFile)
	}
	callBackFuncs.Add(cleanGV)

	// if there are generic things that need to be done, a preStart function could be added here
	// should it be extensive

	// releaseFunc is if the provider starts a vm using a go command
	// and we still need control of it while it is booting until the ready
	// socket is tripped

	releaseCmd, WaitForReady, err := mp.StartVM(mc)
	if err != nil {
		return err
	}

	if WaitForReady == nil {
		return errors.New("no valid WaitForReady function returned")
	}

	// continue check krunkit runnning and wait ready event comming
	if err := WaitForReady(); err != nil {
		return err
	}

	if releaseCmd != nil && releaseCmd() != nil { // some providers can return nil here (hyperv)
		if err := releaseCmd(); err != nil {
			// I think it is ok for a "light" error?
			logrus.Error(err)
		}
	}

	// Actually do nothing
	err = mp.PostStartNetworking(mc, opts.NoInfo)
	if err != nil {
		return err
	}

	// Update state
	stateF := func() (machineDefine.Status, error) {
		return mp.State(mc, true)
	}

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
	}

	// mount the volumes to the VM
	if err := mp.MountVolumesToVM(mc, opts.Quiet); err != nil {
		return err
	}

	// Provider is responsible for waiting
	if mp.UseProviderNetworkSetup() {
		return nil
	}

	machine.WaitAPIAndPrintInfo(
		forwardingState,
		mc.Name,
		forwardSocketPath,
		noInfo,
		mc.HostUser.Rootful,
	)

	return nil
}

func Reset(mps []vmconfigs.VMProvider) error {
	var resetErrors *multierror.Error
	// 注意 machineDefine 是配置模板，不存储数据
	var removeDirs []*machineDefine.MachineDirs

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

// Stop stops the machine as well as supporting binaries/processes
func Stop(mc *vmconfigs.MachineConfig, mp vmconfigs.VMProvider, dirs *machineDefine.MachineDirs, hardStop bool) error {
	// state is checked here instead of earlier because stopping a stopped vm is not considered
	// an error.  so putting in one place instead of sprinkling all over.
	mc.Lock()
	defer mc.Unlock()
	if err := mc.Refresh(); err != nil {
		return fmt.Errorf("reload config: %w", err)
	}

	return stopLocked(mc, mp, dirs, hardStop)
}

// stopLocked stops the machine and expects the caller to hold the machine's lock.
func stopLocked(mc *vmconfigs.MachineConfig, mp vmconfigs.VMProvider, dirs *machineDefine.MachineDirs, hardStop bool) error {
	state, err := mp.State(mc, false)
	if err != nil {
		return err
	}
	// stopping a stopped machine is NOT an error
	if state == machineDefine.Stopped {
		return nil
	}
	if state != machineDefine.Running {
		return machineDefine.ErrWrongState
	}

	// Provider stops the machine
	if err := mp.StopVM(mc, hardStop); err != nil {
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

	// Stop GvProxy and remove PID file
	if !mp.UseProviderNetworkSetup() {
		gvproxyPidFile, err := dirs.RuntimeDir.AppendToNewVMFile("gvproxy.pid")
		if err != nil {
			return err
		}
		if err := gvproxy.CleanupGVProxy(*gvproxyPidFile); err != nil {
			return fmt.Errorf("unable to clean up gvproxy: %w", err)
		}
	}
	// Update last time up
	mc.LastUp = time.Now()
	return mc.Write()
}
