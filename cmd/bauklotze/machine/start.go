package machine

import (
	"bauklotze/cmd/registry"
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/system"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/machine/watcher"
	"bauklotze/pkg/network"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
)

var (
	startCmd = &cobra.Command{
		Use:               "start [options] [MACHINE]",
		Short:             "Start an existing machine",
		Long:              "Start a managed virtual machine ",
		PersistentPreRunE: machinePreRunE, // Get Provider and set workdir if needed
		RunE:              start,
		Args:              cobra.MaximumNArgs(1),
		Example:           `bauklotze machine start`,
	}
	startOpts = define.StartOptions{}
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: startCmd,
		Parent:  machineCmd,
	})

	flags := startCmd.Flags()

	twinPid := ppid
	flags.Int32Var(&startOpts.TwinPid, twinPid, -1, "the pid of PPID")

	ReportUrlFlag := reportUrlFlag
	flags.StringVar(&startOpts.ReportUrl, ReportUrlFlag, "", "Report events to the url")
}

func start(cmd *cobra.Command, args []string) error {
	logrus.Infof("============MachineStart============")
	network.NewReporter(startOpts.ReportUrl)

	if startOpts.TwinPid == -1 {
		mypid := os.Getpid()
		startOpts.TwinPid = int32(mypid)
	}
	if isRunning, err := system.IsProcesSAlive([]int32{startOpts.TwinPid}); !isRunning {
		return err
	}

	ctx, cancelB := context.WithCancel(context.Background())
	defer cancelB()
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case sign := <-signalChan:
			return fmt.Errorf("signal received: %v", sign)
		}
	})

	var err error
	vmName := defaultMachineName
	if len(args) > 0 && len(args[0]) > 0 {
		vmName = args[0]
	}
	dirs, err := env.GetMachineDirs(provider.VMType())
	if err != nil {
		return err
	}
	mc, err := vmconfigs.LoadMachineByName(vmName, dirs)
	if err != nil {
		return err
	}
	watcher.WaitApiServerAndStopMachine(g, ctx, dirs)

	logrus.Infof("starting machine %q\n", vmName)
	network.Reporter.SendEventToOvmJs("start", "vm is staring")

	if err = shim.Start(mc, provider, dirs, startOpts); err != nil {
		return err
	}

	logrus.Infof("Machine %q started successfully\n", vmName)

	if startOpts.TwinPid == -1 {
		mypid := os.Getpid()
		startOpts.TwinPid = int32(mypid)
	}

	watcher.WaitProcessAndStopMachine(g, ctx, startOpts.TwinPid, int32(machine.GlobalPIDs.GetKrunkitPID()), int32(machine.GlobalPIDs.GetGvproxyPID()))
	//watcher.WaitApiServerAndStopMachine(g, ctx, dirs)

	if err = g.Wait(); err != nil {
		logrus.Errorf("%s\n", err.Error())
		logrus.Infof("Try to shutdown the virtualMachine %s gra\n", vmName)
		// TODO: We dont need machine.GlobalPIDs for now
		logrus.Infof("kill krunkit [%d]  and gvproxy [ %d]", machine.GlobalPIDs.GetKrunkitPID(), machine.GlobalPIDs.GetGvproxyPID())
		_ = system.KillProcess(machine.GlobalPIDs.GetGvproxyPID())
		_ = system.KillProcess(machine.GlobalPIDs.GetKrunkitPID())
	}

	gvproxyCmd := machine.GlobalCmds.GetGvproxyCmd()
	logrus.Infof("Waiting for gvproxy to exit,pid [ %d ]", gvproxyCmd.Process.Pid)
	_ = gvproxyCmd.Wait()

	krunCmd := machine.GlobalCmds.GetKrunCmd()
	logrus.Infof("Waiting for krun to exit, pid [ %d ]", krunCmd.Process.Pid)
	_ = krunCmd.Wait()

	return err
}
