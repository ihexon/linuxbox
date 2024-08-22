package machine

import (
	"bauklotze/cmd/registry"
	"bauklotze/pkg/complation"
	"bauklotze/pkg/machine/define"
	provider2 "bauklotze/pkg/machine/provider"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/vmconfigs"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	initCmd = &cobra.Command{
		Use:               "init [options] [NAME]",
		Short:             "Reset and initialize a virtual machine",
		Long:              "Reset and initialize a virtual machine",
		PersistentPreRunE: machinePreRunE,
		RunE:              initMachine,
		Args:              cobra.MaximumNArgs(1), // max positional arguments
		Example:           `podman machine init podman-machine-default`,
		ValidArgsFunction: complation.AutocompleteNone,
	}
	initOpts           = define.InitOptions{}
	defaultMachineName = define.DefaultMachineName
)

type InitOptionalFlags struct {
	UserModeNetworking bool
}

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: initCmd,
		Parent:  machineCmd,
	})

	flags := initCmd.Flags()
	cfg := registry.PodmanConfig()

	flags.BoolVar(
		&initOpts.ReExec,
		"reexec", false,
		"process was rexeced",
	)
	_ = flags.MarkHidden("reexec")

	cpusFlagName := "cpus"
	flags.Uint64Var(
		&initOpts.CPUS,
		cpusFlagName, cfg.ContainersConfDefaultsRO.Machine.CPUs,
		"Number of CPUs",
	)
	_ = initCmd.RegisterFlagCompletionFunc(cpusFlagName, complation.AutocompleteNone)

	diskSizeFlagName := "disk-size"
	flags.Uint64Var(
		&initOpts.DiskSize,
		diskSizeFlagName, cfg.ContainersConfDefaultsRO.Machine.DiskSize,
		"Disk size in GiB",
	)

	memoryFlagName := "memory"
	flags.Uint64VarP(
		&initOpts.Memory,
		memoryFlagName, "m", cfg.ContainersConfDefaultsRO.Machine.Memory,
		"Memory in MiB",
	)
	_ = initCmd.RegisterFlagCompletionFunc(memoryFlagName, complation.AutocompleteNone)

	UsernameFlagName := "username"
	flags.StringVar(&initOpts.Username, UsernameFlagName, cfg.ContainersConfDefaultsRO.Machine.User, "Username used in image")
	_ = initCmd.RegisterFlagCompletionFunc(UsernameFlagName, complation.AutocompleteNone)

	VolumeFlagName := "volume"
	flags.StringArrayVarP(&initOpts.Volumes, VolumeFlagName, "v", cfg.ContainersConfDefaultsRO.Machine.Volumes.Get(), "Volumes to mount, source:target")
	_ = initCmd.RegisterFlagCompletionFunc(VolumeFlagName, complation.AutocompleteDefault)
}

// machinePreRunE: Status ok
func machinePreRunE(c *cobra.Command, args []string) error {
	provider, err := provider2.Get()
	if err != nil {
		return err
	}
	return nil
}

func initMachine(c *cobra.Command, args []string) error {
	initOpts.Name = defaultMachineName
	// Check if machine already exists
	_, exists, err := shim.VMExists(initOpts.Name, []vmconfigs.VMProvider{provider})
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("%s: %w", initOpts.Name, define.ErrVMAlreadyExists)
	}

	err = shim.Init(initOpts, provider)
	if err != nil {
		return err
	}
	return nil
}
