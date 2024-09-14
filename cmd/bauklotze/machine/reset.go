//go:build amd64 || arm64

package machine

import (
	"bauklotze/cmd/bauklotze/validata"
	"bauklotze/cmd/registry"
	provider2 "bauklotze/pkg/machine/provider"
	"bauklotze/pkg/machine/shim"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	resetCmd = &cobra.Command{
		Use:     "reset [options]",
		Short:   "Remove all machines",
		Long:    "Remove all machines, configurations, data, and cached images",
		RunE:    reset,
		Args:    validata.NoArgs,
		Example: `podman machine reset`,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: resetCmd,
		Parent:  machineCmd,
	})
}

func reset(_ *cobra.Command, _ []string) error {
	providers := provider2.GetAll()
	//if err != nil {
	//	return err
	//}
	return shim.Reset(providers)
}

func resetConfirmationMessage(vms []string) {
	fmt.Println("Warning: this command will delete all existing Podman machines")
	fmt.Println("and all of the configuration and data directories for Podman machines")
	fmt.Printf("\nThe following machine(s) will be deleted:\n\n")
	for _, msg := range vms {
		fmt.Printf("%s\n", msg)
	}
}
