package machine

import (
	cmdflags "bauklotze/cmd/bauklotze/flags"
	"bauklotze/cmd/registry"
	cmdproxy "bauklotze/pkg/cliproxy"
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/system"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/machine/watcher"
	"bauklotze/pkg/network"
	"context"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"os"
	"syscall"
)

var (
	startCmd = &cobra.Command{
		Use:   "start [options] [MACHINE]",
		Short: "Start an existing machine",
		Long:  "Start a managed virtual machine ",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logrus.Infof("============startCmd PersistentPreRunE============")
			return machinePreRunE(cmd, args)
		}, // Get Provider and set workdir if needed
		RunE: func(cmd *cobra.Command, args []string) error {
			logrus.Infof("============ startCmd RunE ============")
			return start(cmd, args)
		},
		Args:    cobra.MaximumNArgs(1),
		Example: `bauklotze machine start`,
	}
	startOpts = define.StartOptions{}
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: startCmd,
		Parent:  machineCmd,
	})
}

func start(cmd *cobra.Command, args []string) error {
	var err error

	ppid, _ := cmd.Flags().GetInt32(cmdflags.PpidFlag) // Get PPID from
	logrus.Infof("PID is [%d], PPID is: %d", os.Getpid(), ppid)
	startOpts.CommonOptions.ReportUrl = cmd.Flag(cmdflags.ReportUrlFlag).Value.String()
	startOpts.CommonOptions.PPID = ppid

	if isRunning, err := system.IsProcesSAlive([]int32{ppid}); !isRunning {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	vmName := defaultMachineName
	if len(args) > 0 && len(args[0]) > 0 {
		vmName = args[0]
	}

	dirs, err := env.GetMachineDirs(provider.VMType())
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

	mc, err := vmconfigs.LoadMachineByName(vmName, dirs)
	if err != nil {
		return err
	}

	go func() {
		logrus.Infof("CMDProxy starting...")
		cmdProxyErr := cmdproxy.RunCMDProxy()
		if cmdProxyErr != nil {
			logrus.Errorf("CMDProxy running failed, %v", cmdProxyErr)
		}
		logrus.Warnf("CMDProxy exited")
	}()

	watcher.WaitApiServerAndStopMachine(g, ctx, dirs)

	logrus.Infof("Starting machine %q\n", vmName)
	network.Reporter.SendEventToOvmJs("start", "vm is staring")
	if err = shim.Start(mc, provider, dirs, startOpts); err != nil {
		return err
	}

	logrus.Infof("Machine %q started successfully\n", vmName)
	watcher.WaitProcessAndStopMachine(g, ctx, ppid, int32(machine.GlobalPIDs.GetKrunkitPID()), int32(machine.GlobalPIDs.GetGvproxyPID()))

	err = g.Wait()
	if err != nil {
		logrus.Infof("Do sync in virtualMachine now")
		if sshError := machine.CommonSSHSilent(mc.SSH.RemoteUsername, mc.SSH.IdentityPath, mc.Name, mc.SSH.Port, []string{"sync"}); sshError != nil {
			logrus.Warnf("Failed to sync in virtualMachine: %v", sshError)
		}
		// send SIGTERM to myself, then callback func will get executed
		_ = syscall.Kill(os.Getpid(), syscall.SIGTERM)
	}

	return err
}
