//go:build amd64 || arm64

package machine

import (
	"bauklotze/cmd/bauklotze/validata"
	"bauklotze/cmd/registry"
	"bauklotze/pkg/events"
	"bauklotze/pkg/machine/env"
	provider2 "bauklotze/pkg/machine/provider"
	"bauklotze/pkg/machine/vmconfigs"
	"github.com/spf13/cobra"
	"strings"
)

// TODO: newMachineEvent
func newMachineEvent(status events.Status, event events.Event) {

}

func getMachines(toComplete string) ([]string, cobra.ShellCompDirective) {
	suggestions := []string{}
	provider, err := provider2.Get()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	dirs, err := env.GetMachineDirs(provider.VMType())
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	machines, err := vmconfigs.LoadMachinesInDir(dirs)
	if err != nil {
		cobra.CompErrorln(err.Error())
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	for _, m := range machines {
		if strings.HasPrefix(m.Name, toComplete) {
			suggestions = append(suggestions, m.Name)
		}
	}
	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// autocompleteMachine - Autocomplete machines.
func autocompleteMachine(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return getMachines(toComplete)
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
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
