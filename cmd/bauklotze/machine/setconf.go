//go:build amd64 || arm64

package machine

import (
	"bauklotze/cmd/registry"
	"bauklotze/pkg/completion"
	"bauklotze/pkg/config"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/vmconfigs"
	strongunits "bauklotze/pkg/storage"
	"bauklotze/pkg/system"
	"github.com/spf13/cobra"
)

var (
	setCmd = &cobra.Command{
		Use:               "set [options] [NAME]",
		Short:             "Set a virtual machine setting",
		Long:              "Set an updatable virtual machine setting",
		PersistentPreRunE: machinePreRunE,
		RunE:              setMachine,
		Args:              cobra.MaximumNArgs(1),
		Example:           `machine set --rootful=false`,
	}
)

var (
	setOpts = define.SetOptions{}
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: setCmd,
		Parent:  machineCmd,
	})
	flags := setCmd.Flags()

	cpusFlagName := "cpus"
	flags.Uint64Var(
		&setOpts.CPUs,
		cpusFlagName, 0,
		"Number of CPUs",
	)
	_ = setCmd.RegisterFlagCompletionFunc(cpusFlagName, completion.AutocompleteNone)

	diskSizeFlagName := "disk-size"

	_ = setCmd.RegisterFlagCompletionFunc(diskSizeFlagName, completion.AutocompleteNone)

	memoryFlagName := "memory"
	flags.Uint64VarP(
		&setOpts.Memory,
		memoryFlagName, "m", 0,
		"Memory in MiB",
	)
	_ = setCmd.RegisterFlagCompletionFunc(memoryFlagName, completion.AutocompleteNone)

	slice := config.NewSlice([]string{})
	volumeFlagName := "volume"
	flags.StringArrayVarP(
		&setOpts.Volumes,
		volumeFlagName, "v", slice.Get(),
		"Volume to be mounted in the VM",
	)
	_ = setCmd.RegisterFlagCompletionFunc(volumeFlagName, completion.AutocompleteNone)
}

func setMachine(cmd *cobra.Command, args []string) error {
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

	if cmd.Flags().Changed("cpus") {
		setOpts.CPUs = setOpts.CPUs
	}
	if cmd.Flags().Changed("memory") {
		newMemory := strongunits.MiB(setOpts.Memory)
		if err := system.CheckMaxMemory(newMemory); err != nil {
			return err
		}
	}

	// At this point, we have the known changed information, etc
	// Walk through changes to the providers if they need them
	return shim.Set(mc, provider, setOpts)
}
