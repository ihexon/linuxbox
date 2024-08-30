package machine

import (
	"bauklotze/cmd/registry"
	"bauklotze/pkg/completion"
	"bauklotze/pkg/events"
	"bauklotze/pkg/machine/define"
	provider2 "bauklotze/pkg/machine/provider"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/vmconfigs"
	strongunits "bauklotze/pkg/storage"
	"fmt"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/cobra"
	"os"
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
		ValidArgsFunction: completion.AutocompleteNone,
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
	cfg := registry.OvmConfig()

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
	_ = initCmd.RegisterFlagCompletionFunc(cpusFlagName, completion.AutocompleteNone)

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
	_ = initCmd.RegisterFlagCompletionFunc(memoryFlagName, completion.AutocompleteNone)

	UsernameFlagName := "username"
	flags.StringVar(&initOpts.Username, UsernameFlagName, cfg.ContainersConfDefaultsRO.Machine.User, "Username used in image")
	_ = initCmd.RegisterFlagCompletionFunc(UsernameFlagName, completion.AutocompleteNone)

	VolumeFlagName := "volume"
	flags.StringArrayVarP(&initOpts.Volumes, VolumeFlagName, "v", cfg.ContainersConfDefaultsRO.Machine.Volumes.Get(), "Volumes to mount, source:target")
	_ = initCmd.RegisterFlagCompletionFunc(VolumeFlagName, completion.AutocompleteDefault)

	ImageFlagName := "image"
	flags.StringVar(&initOpts.Image, ImageFlagName, cfg.ContainersConfDefaultsRO.Machine.Image, "Bootable image for machine")
	_ = initCmd.RegisterFlagCompletionFunc(ImageFlagName, completion.AutocompleteDefault)
}

// machinePreRunE: Status ok
func machinePreRunE(c *cobra.Command, args []string) error {
	var err error = nil
	provider, err = provider2.Get()
	if err != nil {
		return err
	}
	return nil
}

func initMachine(cmd *cobra.Command, args []string) error {
	initOpts.Name = defaultMachineName
	// Check if machine already exists
	// In macos_arm64 shim.VMExist always false
	_, exists, err := shim.VMExists(initOpts.Name, []vmconfigs.VMProvider{provider})
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("%s: %w", initOpts.Name, define.ErrVMAlreadyExists)
	}

	for idx, vol := range initOpts.Volumes {
		initOpts.Volumes[idx] = os.ExpandEnv(vol)
	}

	if cmd.Flags().Changed("memory") {
		if err := checkMaxMemory(strongunits.MiB(initOpts.Memory)); err != nil {
			return err
		}
	}

	err = shim.Init(initOpts, provider)
	if err != nil {
		return err
	}
	newMachineEvent(events.Init, events.Event{Name: initOpts.Name})
	return nil
}

// checkMaxMemory gets the total system memory and compares it to the variable.  if the variable
// is larger than the total memory, it returns an error
func checkMaxMemory(newMem strongunits.MiB) error {
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	if total := strongunits.B(memStat.Total); strongunits.B(memStat.Total) < newMem.ToBytes() {
		return fmt.Errorf("requested amount of memory (%d MB) greater than total system memory (%d MB)", newMem, total)
	}
	return nil
}
