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
	"bauklotze/pkg/notifyexit"
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"os"
	"time"
)

var (
	startCmd = &cobra.Command{
		Use:               "start [options] [MACHINE]",
		Short:             "Start an existing machine",
		Long:              "Start a managed virtual machine ",
		PersistentPreRunE: machinePreRunE, // Get Provider and set workdir if needed
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			network.Reporter.SendEventToOvmJs("exit", "")
			return nil
		},
		RunE:    start,
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

	flags := startCmd.Flags()

	twinPid := ppid
	flags.Int32Var(&startOpts.TwinPid, twinPid, -1, "the pid of PPID")

	ReportUrlFlag := reportUrlFlag
	flags.StringVar(&startOpts.ReportUrl, ReportUrlFlag, "", "Report events to the url")
}

func start(cmd *cobra.Command, args []string) error {
	network.NewReporter(startOpts.ReportUrl)
	ctxA, cancelA := context.WithCancel(context.Background())
	defer cancelA()
	g, ctxA := errgroup.WithContext(ctxA)
	// If not specified PPID, use the current process id as the parent process id
	if startOpts.TwinPid == -1 {
		mypid := os.Getpid()
		startOpts.TwinPid = int32(mypid)
	}

	g.Go(func() error {
		for {
			select {
			case <-ctxA.Done():
				return context.Cause(ctxA)
			default:
			}
			if isRunning, err := system.IsProcesSAlive([]int32{startOpts.TwinPid}); !isRunning {
				return err
			}
			time.Sleep(300 * time.Millisecond)
		}
	})

	go func() {
		if err := g.Wait(); err != nil && !(errors.Is(err, context.Canceled)) {
			logrus.Errorf("%s\n", err.Error())
			network.Reporter.SendEventToOvmJs("error", err.Error())
			registry.SetExitCode(1)
			notifyexit.NotifyExit(registry.GetExitCode())
		}
	}()

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

	logrus.Infof("starting machine %q\n", vmName)
	network.Reporter.SendEventToOvmJs("start", "vm is staring")

	if err = shim.Start(mc, provider, dirs, startOpts); err != nil {
		return err
	}

	logrus.Infof("Machine %q started successfully\n", vmName)

	ctxB, cancelB := context.WithCancel(context.Background())
	defer cancelB()
	cancelA() // Cancel the first context because we do not need that anymore
	g, ctxB = errgroup.WithContext(ctxB)

	if startOpts.TwinPid == -1 {
		mypid := os.Getpid()
		startOpts.TwinPid = int32(mypid)
	}

	watcher.WaitProcessAndStopMachine(g, ctxB, startOpts.TwinPid, int32(machine.GlobalPIDs.GetKrunkitPID()), int32(machine.GlobalPIDs.GetGvproxyPID()))
	watcher.WaitApiServerAndStopMachine(g, ctxB, dirs)

	if err := g.Wait(); err != nil {
		logrus.Errorf("%s\n", err.Error())
		_ = system.KillProcess(machine.GlobalPIDs.GetGvproxyPID())
		_ = system.KillProcess(machine.GlobalPIDs.GetKrunkitPID())
		return err
	}

	return err
}
