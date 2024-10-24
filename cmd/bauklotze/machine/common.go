//go:build amd64 || arm64

package machine

import (
	cmdflags "bauklotze/cmd/bauklotze/flags"
	"bauklotze/cmd/bauklotze/validata"
	"bauklotze/cmd/registry"
	"bauklotze/pkg/machine/env"
	provider2 "bauklotze/pkg/machine/provider"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/network"
	"github.com/spf13/cobra"
)

var provider vmconfigs.VMProvider

func machinePreRunE(cmd *cobra.Command, args []string) error {
	var err error = nil
	network.NewReporter(commonOpts.ReportUrl) // Initialize the network.NewReporter
	d, _ := cmd.Flags().GetString(cmdflags.WorkspaceFlag)
	env.InitCustomHomeEnvOnce(d)

	provider, err = provider2.Get()
	if err != nil {
		return err
	}

	return nil
}

var machineCmd = &cobra.Command{
	Use:   "machine",
	Short: "Manage a virtual machine",
	Long:  "Manage a virtual machine. Virtual machines are used to run OVM.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	PersistentPostRunE: closeMachineEvents,
	RunE:               validata.SubCommandExists,
}

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{Command: machineCmd})
}

func closeMachineEvents(cmd *cobra.Command, _ []string) error {
	return nil
}
