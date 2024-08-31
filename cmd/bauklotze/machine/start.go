package machine

import (
	"bauklotze/cmd/registry"
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/vmconfigs"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	startCmd = &cobra.Command{
		Use:               "start [options] [MACHINE]",
		Short:             "Start an existing machine",
		Long:              "Start a managed virtual machine ",
		PersistentPreRunE: machinePreRunE,
		RunE:              start,
		Args:              cobra.MaximumNArgs(1),
		Example:           `podman machine start podman-machine-default`,
		ValidArgsFunction: autocompleteMachine,
	}
	startOpts = machine.StartOptions{}
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: startCmd,
		Parent:  machineCmd,
	})

	flags := startCmd.Flags()
	noInfoFlagName := "no-info"
	flags.BoolVar(&startOpts.NoInfo, noInfoFlagName, false, "Suppress informational tips")

	quietFlagName := "quiet"
	flags.BoolVarP(&startOpts.Quiet, quietFlagName, "q", false, "Suppress machine starting status output")
}

func start(_ *cobra.Command, args []string) error {
	var (
		err error
	)
	startOpts.NoInfo = startOpts.Quiet || startOpts.NoInfo
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

	if !startOpts.Quiet {
		fmt.Printf("Starting machine %q\n", vmName)
	}
	// TODO
	if err := shim.Start(mc, provider, dirs, startOpts); err != nil {
		return err
	}
	fmt.Printf("Machine %q started successfully\n", vmName)
	return nil

}
