//go:build amd64 || arm64

package machine

import (
	"bauklotze/cmd/bauklotze/validata"
	"bauklotze/cmd/registry"
	"bauklotze/pkg/events"
	"bauklotze/pkg/machine/vmconfigs"
	"github.com/spf13/cobra"
)

// TODO: newMachineEvent
func newMachineEvent(status events.Status, event events.Event) {

}

var (
	provider vmconfigs.VMProvider
)

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
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: machineCmd,
	})
}

func closeMachineEvents(cmd *cobra.Command, _ []string) error {
	return nil
}
