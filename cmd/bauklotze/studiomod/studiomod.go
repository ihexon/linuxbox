package studiomod

import (
	"bauklotze/cmd/bauklotze/validata"
	"bauklotze/cmd/registry"
	"github.com/spf13/cobra"
)

var studiomodCmd = &cobra.Command{
	Use:   "studiomod",
	Short: "Manage a virtual machine in studiomod",
	Long:  "Manage a virtual machine. Virtual machines are used to run OVM.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	PersistentPostRunE: closeMachineEvents,
	RunE:               validata.SubCommandExists,
	Hidden:             false,
}

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: studiomodCmd,
	})
}

func closeMachineEvents(cmd *cobra.Command, _ []string) error {
	return nil
}
