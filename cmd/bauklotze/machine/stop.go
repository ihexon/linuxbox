//go:build amd64 || arm64

package machine

import (
	"bauklotze/cmd/registry"
	"bauklotze/pkg/events"
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/vmconfigs"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	stopCmd = &cobra.Command{
		Use:               "stop [MACHINE]",
		Short:             "Stop an existing machine",
		Long:              "Stop a managed virtual machine ",
		PersistentPreRunE: machinePreRunE,
		RunE:              stop,
		Args:              cobra.MaximumNArgs(1),
		Example:           `podman machine stop podman-machine-default`,
		ValidArgsFunction: autocompleteMachine,
	}
	stopOpts = machine.StopOptions{}
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: stopCmd,
		Parent:  machineCmd,
	})

}

// TODO  Name shouldn't be required, need to create a default vm
func stop(cmd *cobra.Command, args []string) error {
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

	if err := shim.Stop(mc, provider, dirs, false); err != nil {
		return err
	}

	fmt.Printf("Machine %q stopped successfully\n", vmName)

	// TODO: Scan Event Socks dir and send event to all socks file
	err = NewMachineEvent(events.Stop, "stopped", mc)
	if err != nil {
		logrus.Warnf("Send event failed: %s", err.Error())
	}
	return err
}
