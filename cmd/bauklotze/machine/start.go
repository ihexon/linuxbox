package machine

import (
	"bauklotze/cmd/registry"
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/system"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
		ValidArgsFunction: autocompleteMachine,
	}
	startOpts = define.StartOptions{
		WaitAndStop: false,
	}
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: startCmd,
		Parent:  machineCmd,
	})
	flags := startCmd.Flags()

	waitAndStop := "waitAndStop"
	flags.BoolVarP(&startOpts.WaitAndStop,
		waitAndStop,
		"",
		false,
		"When any of ppid, gvproxy, and krunkit got exit, STOP the virtual Machine")
	flags.MarkHidden(waitAndStop)

	twinPid := "twinpid"
	flags.IntVar(&startOpts.TwinPid, twinPid, -1, "the pid of PPID")
	flags.MarkHidden(twinPid)
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

	logrus.Infof("Machine %q started successfully\n", vmName)

	if startOpts.WaitAndStop || (startOpts.TwinPid != -1) {
		logrus.Infof("Waiting PPID[%d] exited then stop the machine\n", startOpts.TwinPid)
		return WaitingAndKillProcess(cmd, args,
			startOpts.TwinPid,
			machine.GlobalPIDs.GetKrunkitPID(),
			machine.GlobalPIDs.GetGvproxyPID())
	}

	return nil
}

func WaitingAndKillProcess(cmd *cobra.Command, args []string, ovmppid, krunkit, gvproxy int) error {
	var err error
	somethingWrong := make(chan bool)
	go func() {
		for {
			if ovmppid != -1 {
				if ok := system.IsProcessAlive(ovmppid); !ok {
					somethingWrong <- true
					return
				}
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

	select {
	case exited := <-somethingWrong:
		if exited == true {
			return stop(cmd, args)
		}
	}
	return err
}

// WaitingAndKillProcessV2 : Not testing, maybe do not work as I expect
func WaitingAndKillProcessV2(cmd *cobra.Command, args []string, ovmppid, krunkit, gvproxy int) error {
	somethingWrong := make(chan bool)
	go func() {
		for {
			if ovmppid != -1 && !system.IsProcessAlive(ovmppid) {
				somethingWrong <- true
				return
			}
			if system.CheckProcessRunning("Krunkit", krunkit) != nil || system.CheckProcessRunning("GVproxy", gvproxy) != nil {
				somethingWrong <- true
				return
			}
			time.Sleep(400 * time.Millisecond)
		}
	}()

	if <-somethingWrong {
		logrus.Infof("ppid:[%d] exited ! stop machine now...", ovmppid)
		//return stop(cmd, args)
	}
	return nil
}
