package machine

import (
	cmdflags "bauklotze/cmd/bauklotze/flags"
	"bauklotze/cmd/registry"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/system"
	"bauklotze/pkg/machine/vmconfigs"
	system2 "bauklotze/pkg/system"
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
	RegexError    = fmt.Errorf("names must match [a-zA-Z0-9][a-zA-Z0-9_.-]*: %w", ErrInvalidArg)
	ErrInvalidArg = errors.New("invalid argument")
)

var (
	initCmd = &cobra.Command{
		Use:   "init [options] [NAME]",
		Short: "initialize a virtual machine",
		Long:  "initialize a virtual machine",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logrus.Infof("============initCmd PersistentPreRunE============")
			return machinePreRunE(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logrus.Infof("============ initCmd RunE ============")
			return initMachine(cmd, args)
		},
		Args:    cobra.MaximumNArgs(1), // max positional arguments
		Example: `machine init default`,
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

	cpusFlagName := cmdflags.CpusFlag
	flags.Uint64Var(
		&initOpts.CPUS,
		cpusFlagName, cfg.ContainersConfDefaultsRO.Machine.CPUs,
		"Number of CPUs",
	)

	memoryFlagName := cmdflags.MemoryFlag
	flags.Uint64VarP(
		&initOpts.Memory,
		memoryFlagName, "m", cfg.ContainersConfDefaultsRO.Machine.Memory,
		"Memory in MiB",
	)

	VolumeFlagName := cmdflags.VolumeFlag
	flags.StringArrayVarP(&initOpts.Volumes, VolumeFlagName, "v", cfg.ContainersConfDefaultsRO.Machine.Volumes.Get(), "Volumes to mount, source:target")

	BootImageName := cmdflags.BootImageFlag
	flags.StringVar(&initOpts.Images.BootableImage, BootImageName, cfg.ContainersConfDefaultsRO.Machine.Image, "Bootable image for machine")
	_ = initCmd.MarkFlagRequired(BootImageName)

	BootImageVersion := cmdflags.BootVersionFlag
	flags.StringVar(&initOpts.ImageVersion.BootableImageVersion, BootImageVersion, cfg.ContainersConfDefaultsRO.Machine.Image, "Boot version field")
	initCmd.MarkFlagRequired(BootImageVersion)

	DataImageVersion := cmdflags.DataVersionFlag
	flags.StringVar(&initOpts.ImageVersion.DataDiskVersion, DataImageVersion, "", "Data version field")
	initCmd.MarkFlagRequired(DataImageVersion)

}

func initMachine(cmd *cobra.Command, args []string) error {
	var err error
	// TODO Use ctx to get some parameters would be nice, also using ctx to control the lifecycle init()
	//ctx := cmd.Context()
	//ctx, cancel := context.WithCancelCause(ctx)
	//logrus.Infof("cmd.Context().Value(\"commonOpts\") --> %v", ctx.Value("commonOpts"))

	ppid, _ := cmd.Flags().GetInt32(cmdflags.PpidFlag) // Get PPID from
	logrus.Infof("PID is [%d], PPID is: %d", os.Getpid(), ppid)

	initOpts.CommonOptions.ReportUrl = cmd.Flag(cmdflags.ReportUrlFlag).Value.String()
	initOpts.CommonOptions.PPID = ppid

	// TODO Continue to check the ppid alive
	// First check the parent process is alive once
	if isRunning, err := system.IsProcesSAlive([]int32{ppid}); !isRunning {
		return err
	}
	logrus.Infof("Initialize machine name %s", defaultMachineName)
	initOpts.Name = defaultMachineName
	if len(args) > 0 {
		if len(args[0]) > cmdflags.MaxMachineNameSize {
			return fmt.Errorf("machine name %q must be %d characters or less", args[0], cmdflags.MaxMachineNameSize)
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

	dataDir, err := env.DataDirPrefix() // ${BauklotzeHomePath}/data
	if err != nil {
		return fmt.Errorf("can not get Data dir %v", err)
	}

	dataDisk := filepath.Join(dataDir, "external_disk", initOpts.Name, "data.raw") // ${BauklotzeHomePath}/data/{MachineName}/data.raw
	initOpts.Images.DataDisk = dataDisk

	var (
		updateBootableImage = true
		updateExternalDisk  = true
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
	case oldMc.DataDiskVersion == initOpts.ImageVersion.DataDiskVersion: // If old version != given version
		updateExternalDisk = false
	default:
		updateExternalDisk = true
	}

	if updateExternalDisk {
		if initOpts.Images.DataDisk != "" {
			logrus.Infof("Recreate data disk: %s", initOpts.Images.DataDisk)
			err = system2.CreateAndResizeDisk(initOpts.Images.DataDisk, strongunits.GiB(100))
			if err != nil {
				return err
			}
		}
	} else {
		logrus.Infof("Skip initialize data disk.")
	}

	if !updateBootableImage {
		logrus.Infof("skip initialize virtual machine")
		return nil
	}

	for idx, vol := range initOpts.Volumes {
		initOpts.Volumes[idx] = os.ExpandEnv(vol)
	}

	if err = systemResourceCheck(cmd); err != nil {
		return err
	}

	logrus.Infof("Initialize virtual machine %s with %s", initOpts.Name, initOpts.Images.BootableImage)
	err = shim.Init(initOpts, provider)
	if err != nil {
		return err
	}
	return nil
}

func systemResourceCheck(cmd *cobra.Command) error {
	var err error
	if cmd.Flags().Changed("memory") {
		if err = system2.CheckMaxMemory(strongunits.MiB(initOpts.Memory)); err != nil {
			logrus.Errorf("Can not allocate the memory size %s", initOpts.Memory)
			return err
		}
	}

	// Krun limited max cpus core to 8
	if cmd.Flags().Changed(cmdflags.CpusFlag) {
		if initOpts.CPUS > cmdflags.KrunMaxCpus || initOpts.CPUS < 1 {
			return fmt.Errorf("can not allocate the CPU size %d", initOpts.CPUS)
		}
	}

	return err
}
