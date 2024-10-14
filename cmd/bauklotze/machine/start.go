package machine

import (
	"bauklotze/cmd/registry"
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/system"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/network"
	"context"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"net/url"
	"os"
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

	reportUrl := reportUrl
	flags.StringVar(&startOpts.ReportUrl, reportUrl, "", "Report events to the url")
}

func start(cmd *cobra.Command, args []string) error {
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
	if err = shim.Start(mc, provider, dirs, startOpts); err != nil {
		return err
	}

	if startOpts.ReportUrl != "" {
		connCtx, err := network.NewConnection(startOpts.ReportUrl)
		if err != nil {
			logrus.Errorf("Failed to connect to %q: %v\n", startOpts.ReportUrl, err)
		}
		connCtx.UrlParameter = url.Values{
			"event":   []string{"running"},
			"message": []string{"ready"},
		}
		// ? Should I return error ?
		_, err = connCtx.DoRequest("GET", "/notify", nil)
		if err != nil {
			logrus.Warnf("Failed to notify %q: %v\n", startOpts.ReportUrl, err)
		}
	}

	logrus.Infof("Machine %q started successfully\n", vmName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	mypid := os.Getpid()
	startOpts.TwinPid = int32(mypid)
	system.WaitProcessAndStopMachine(g, ctx, startOpts.TwinPid, int32(machine.GlobalPIDs.GetKrunkitPID()), int32(machine.GlobalPIDs.GetGvproxyPID()))
	system.WaitApiServerAndStopMachine(g, ctx, dirs)

	if err := g.Wait(); err != nil {
		logrus.Errorf("%s\n", err.Error())
		return stop(cmd, args)
	}
	return err
}
