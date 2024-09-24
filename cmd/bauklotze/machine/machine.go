//go:build amd64 || arm64

package machine

import (
	"bauklotze/cmd/bauklotze/validata"
	"bauklotze/cmd/registry"
	"bauklotze/pkg/machine/env"
	provider2 "bauklotze/pkg/machine/provider"
	"bauklotze/pkg/machine/vmconfigs"
	"github.com/spf13/cobra"
	"strings"
)

//func NewMachineEvent(event events.Status, msg string, mc *vmconfigs.MachineConfig) error {
//	sockPath := mc.EvtSockPath.GetPath()
//	if sockPath == "" {
//		// Do Nothing
//		return nil
//	}
//
//	fileInfo, err := os.Stat(mc.EvtSockPath.GetPath())
//	if err != nil {
//		return err
//	}
//
//	if fileInfo.IsDir() {
//		// Do scan dir....
//		return nil
//	}
//
//	if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
//		if stat.Mode&syscall.S_IFMT == syscall.S_IFSOCK {
//			network.SendEventToOvmJs(event, msg, mc.EvtSockPath.GetPath())
//		}
//	}
//
//	return nil
//}

var provider vmconfigs.VMProvider

func machinePreRunE(cmd *cobra.Command, args []string) error {
	var err error = nil
	d, _ := cmd.Flags().GetString("workdir")
	env.InitCustomHomeEnv(d)

	provider, err = provider2.Get()
	if err != nil {
		return err
	}
	return nil
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
