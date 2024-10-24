package machine

import (
	cmdflags "bauklotze/cmd/bauklotze/flags"
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

	ppid := cmdflags.PpidFlag // Default is -1
	flags.Int32Var(&startOpts.PPID, ppid, -1, "the pid of PPID")

	ReportUrlFlag := cmdflags.ReportUrlFlag
	flags.StringVar(&commonOpts.ReportUrl, ReportUrlFlag, "", "Report events to the url")
}

func start(cmd *cobra.Command, args []string) error {
	var err error
	logrus.Infof("============MachineStart============")
	if startOpts.PPID == -1 {
		startOpts.PPID, err = system.GetPPID(int32(os.Getpid()))
		if err != nil {
			return fmt.Errorf("failed to get parent pid: %w", err)
		} else {
			logrus.Infof("The parent pid is: %d", startOpts.PPID)
		}
	}

	if isRunning, err := system.IsProcesSAlive([]int32{startOpts.PPID}); !isRunning {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case sign := <-signalChan:
			return fmt.Errorf("Signal received: %v", sign)
		}
	})

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

	watcher.WaitProcessAndStopMachine(g, ctx, startOpts.PPID, int32(machine.GlobalPIDs.GetKrunkitPID()), int32(machine.GlobalPIDs.GetGvproxyPID()))

	if err = g.Wait(); err != nil {
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
