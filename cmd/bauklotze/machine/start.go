package machine

import (
	"bauklotze/cmd/registry"
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/network"
	"bauklotze/pkg/system"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"time"
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
		_, _ = connCtx.DoRequest("GET", "/notify", nil)
	}

	logrus.Infof("Machine %q started successfully\n", vmName)

	// If user do not --twinpid, get my PPID
	if startOpts.TwinPid == -1 {
		startOpts.TwinPid, err = system.GetMyPPID()
		if err != nil {
			return err
		}
	}

	return WaiteAndStopMachine(
		startOpts,
		args,
		machine.GlobalPIDs.GetKrunkitPID(),
		machine.GlobalPIDs.GetGvproxyPID(),
	)
}

func WaiteAndStopMachine(startOpts define.StartOptions, args []string, krunkit, gvproxy int) error {
	if startOpts.TwinPid != -1 {
		logrus.Infof("Waiting PPID[%d] exited then stop the machine\n", startOpts.TwinPid)
		return waiteAndStopMachine(args, int(startOpts.TwinPid), krunkit, gvproxy)
	}
	return nil
}

func waiteAndStopMachine(args []string, ovmppid, krunkit, gvproxy int) error {
	var err error
	somethingWrong := make(chan bool)
	go func() {
		for {
			if ovmppid != -1 && !system.IsProcessAlive(ovmppid) {
				somethingWrong <- true
				return
			}
			// Notice the CheckProcessRunning is a NO-BLOCK function
			if err := system.CheckProcessRunning("KRunkit", krunkit); err != nil {
				somethingWrong <- true
				return
			}
			// Notice the CheckProcessRunning is a NO-BLOCK function
			if err := system.CheckProcessRunning("GVProxy", gvproxy); err != nil {
				somethingWrong <- true
				return
			}
			// lets poll status every half second
			time.Sleep(400 * time.Millisecond)
		}
	}()

	if <-somethingWrong {
		return stop(nil, args)
	}
	return err
}
