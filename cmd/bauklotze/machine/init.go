package machine

import (
	"bauklotze/cmd/registry"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/system"
	"errors"
	"fmt"
	"github.com/containers/common/pkg/strongunits"
	"github.com/containers/storage/pkg/regexp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var (
	NameRegex     = regexp.Delayed("^[a-zA-Z0-9][a-zA-Z0-9_.-]*$")
	RegexError    = fmt.Errorf("names must match [a-zA-Z0-9][a-zA-Z0-9_.-]*: %w", ErrInvalidArg) // nolint:revive // This lint is new and we do not want to break the API.
	ErrInvalidArg = errors.New("invalid argument")
)

var (
	initCmd = &cobra.Command{
		Use:               "init [options] [NAME]",
		Short:             "initialize a virtual machine",
		Long:              "initialize a virtual machine",
		PersistentPreRunE: machinePreRunE,
		RunE:              initMachine,
		Args:              cobra.MaximumNArgs(1), // max positional arguments
		Example:           `machine init`,
	}
	initOpts = define.InitOptions{
		Username: define.DefaultUserInGuest,
	}
	defaultMachineName = define.DefaultMachineName
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: initCmd,
		Parent:  machineCmd,
	})

	// Calculate the default configuration
	// CPU, MEMORY, etc.
	// OvmInitConfig() configures the memory/CPU/disk size/external mount points for the virtual machine.
	// These configurations will be written to the machine's JSON file for persistence.
	cfg := registry.OvmInitConfig()
	flags := initCmd.Flags()

	cpusFlagName := cpus
	flags.Uint64Var(
		&initOpts.CPUS,
		cpusFlagName, cfg.ContainersConfDefaultsRO.Machine.CPUs,
		"Number of CPUs",
	)

	memoryFlagName := memory
	flags.Uint64VarP(
		&initOpts.Memory,
		memoryFlagName, "m", cfg.ContainersConfDefaultsRO.Machine.Memory,
		"Memory in MiB",
	)

	VolumeFlagName := volume
	flags.StringArrayVarP(&initOpts.Volumes, VolumeFlagName, "v", cfg.ContainersConfDefaultsRO.Machine.Volumes.Get(), "Volumes to mount, source:target")

	BootImageName := bootImage
	flags.StringVar(&initOpts.Images.BootableImage, BootImageName, cfg.ContainersConfDefaultsRO.Machine.Image, "Bootable image for machine")

	BootImageVersion := bootVersion
	flags.StringVar(&initOpts.ImageVersion.BootableImageVersion, BootImageVersion, cfg.ContainersConfDefaultsRO.Machine.Image, "Bootable image for machine")

	sendEventToEndpoint := reportUrl
	flags.StringVar(&initOpts.SendEvt, sendEventToEndpoint, "", "send events to somewhere")
}

func initMachine(cmd *cobra.Command, args []string) error {
	initOpts.Name = defaultMachineName
	if len(args) > 0 {
		if len(args[0]) > maxMachineNameSize {
			return fmt.Errorf("machine name %q must be %d characters or less", args[0], maxMachineNameSize)
		}
		initOpts.Name = args[0]
		if !NameRegex.MatchString(initOpts.Name) {
			return fmt.Errorf("invalid name %q: %w", initOpts.Name, RegexError)
		}
	}

	oldMc, _, err := shim.VMExists(initOpts.Name, []vmconfigs.VMProvider{provider})
	if err != nil {
		return err
	}

	d, err := env.DataDirPrefix() // ${BauklotzeHomePath}/data
	if err != nil {
		return fmt.Errorf("can not get Data dir %v", err)
	}
	d = filepath.Join(d, "external_disk", initOpts.Name, "data.raw") // ${BauklotzeHomePath}/data
	initOpts.Images.ExternalDisk = d

	var (
		updateBootableImage bool = true
		updateExternalDisk  bool = true
	)

	switch {
	case oldMc == nil: // If machine not initialize before
		updateBootableImage = true
	case oldMc.ImageVersion == initOpts.ImageVersion.BootableImageVersion: // If old version != given version
		updateBootableImage = false
	default:
		updateBootableImage = true
	}

	switch {
	case oldMc == nil: // If machine not initialize before
		updateExternalDisk = true
	case oldMc.ExternalDiskVersion == initOpts.ImageVersion.ExternalDiskVersion: // If old version != given version
		updateExternalDisk = false
	default:
		updateExternalDisk = true
	}

	if updateExternalDisk {
		if initOpts.Images.ExternalDisk != "" {
			logrus.Infof("Recreate external disk: %s", initOpts.Images.ExternalDisk)
			err = system.CreateAndResizeDisk(initOpts.Images.ExternalDisk, strongunits.GiB(100))
			if err != nil {
				return err
			}
		}
	} else {
		logrus.Infof("Skip initialize external disk.")
	}

	if !updateBootableImage {
		return fmt.Errorf("skip initialize virtualMachine")
	}

	for idx, vol := range initOpts.Volumes {
		initOpts.Volumes[idx] = os.ExpandEnv(vol)
	}

	// The allocate virtual memory can not bigger than physic virtual memory
	if cmd.Flags().Changed("memory") {
		if err := system.CheckMaxMemory(strongunits.MiB(initOpts.Memory)); err != nil {
			logrus.Infof("Can not allocate the memory size %s", initOpts.Memory)
			return err
		}
	}

	err = shim.Init(initOpts, provider)
	if err != nil {
		return err
	}

	return nil
}
