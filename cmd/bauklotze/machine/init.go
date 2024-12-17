//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package machine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	cmdflags "bauklotze/cmd/bauklotze/flags"
	"bauklotze/cmd/registry"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/system"
	"bauklotze/pkg/machine/vmconfigs"
	system2 "bauklotze/pkg/system"

	"github.com/containers/common/pkg/strongunits"
	"github.com/containers/storage/pkg/regexp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	NameRegex     = regexp.Delayed("^[a-zA-Z0-9][a-zA-Z0-9_.-]*$")
	ErrRegex      = fmt.Errorf("names must match [a-zA-Z0-9][a-zA-Z0-9_.-]*: %w", ErrInvalidArg)
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
		Example:           `machine init default`,
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
	flags.StringVar(&initOpts.ImagesStruct.BootableImage, BootImageName, cfg.ContainersConfDefaultsRO.Machine.Image, "Bootable image for machine")
	_ = initCmd.MarkFlagRequired(BootImageName)

	BootImageVersion := cmdflags.BootVersionFlag
	flags.StringVar(&initOpts.ImageVerStruct.BootableImageVersion, BootImageVersion, cfg.ContainersConfDefaultsRO.Machine.Image, "Boot version field")
	_ = initCmd.MarkFlagRequired(BootImageVersion)

	DataImageVersion := cmdflags.DataVersionFlag
	flags.StringVar(&initOpts.ImageVerStruct.DataDiskVersion, DataImageVersion, "", "Data version field")
	_ = initCmd.MarkFlagRequired(DataImageVersion)
}

const (
	initfsDir  = "/tmp/initfs"
	initfsArgs = initfsDir + ":" + initfsDir
)

func initMachine(cmd *cobra.Command, args []string) error {
	var err error
	// TODO Use ctx to get some parameters would be nice, also using ctx to control the lifecycle init()
	// ctx := cmd.Context()
	// ctx, cancel := context.WithCancelCause(ctx)
	// logrus.Infof("cmd.Context().Value(\"commonOpts\") --> %v", ctx.Value("commonOpts"))

	ppid, _ := cmd.Flags().GetInt32(cmdflags.PpidFlag) // Get PPID from
	logrus.Infof("PID is [ %d ], watching PPID: [ %d ]", os.Getpid(), ppid)

	initOpts.CommonOptions.ReportURL = cmd.Flag(cmdflags.ReportURLFlag).Value.String()
	initOpts.CommonOptions.PPID = ppid
	// Ignition scripts placed in /tmp/initfs will be executed by the ovmounter service
	initOpts.Volumes = append(initOpts.Volumes, initfsArgs)

	// TODO Continue to check the ppid alive
	// First check the parent process is alive once
	if isRunning, err := system.IsProcesSAlive([]int32{ppid}); !isRunning {
		return fmt.Errorf("parent process %d is not alive: %w", ppid, err)
	}

	initOpts.Name = defaultMachineName
	if len(args) > 0 {
		if len(args[0]) > cmdflags.MaxMachineNameSize {
			return fmt.Errorf("machine name %q must be %d characters or less", args[0], cmdflags.MaxMachineNameSize)
		}
		initOpts.Name = args[0]
		if !NameRegex.MatchString(initOpts.Name) {
			return fmt.Errorf("invalid name %q: %w", initOpts.Name, ErrRegex)
		}
	}

	oldMc, _, err := shim.VMExists(initOpts.Name, []vmconfigs.VMProvider{provider})
	if err != nil {
		return fmt.Errorf("check machine exists error: %w", err)
	}

	// update machine configure

	dataDir, err := env.DataDirPrefix() // ${BauklotzeHomePath}/data
	if err != nil {
		return fmt.Errorf("can not get Data dir %w", err)
	}

	dataDisk := filepath.Join(dataDir, "external_disk", initOpts.Name, "data.raw") // ${BauklotzeHomePath}/data/{MachineName}/data.raw
	initOpts.ImagesStruct.DataDisk = dataDisk

	// Default do not update anything
	var (
		updateBootableImage = false
		updateExternalDisk  = false
	)

	switch {
	case oldMc == nil: // If machine not initialize before, mark updateBootableImage=true  && updateExternalDisk=true
		updateExternalDisk = true
	case oldMc.DataDiskVersion != initOpts.ImageVerStruct.DataDiskVersion: // If old DataDisk version != given DataDisk version
		updateExternalDisk = true
	}

	switch {
	case oldMc == nil: // If machine not initialize before, mark updateBootableImage=true  && updateExternalDisk=true
		updateBootableImage = true
	case oldMc.BootableDiskVersion != initOpts.ImageVerStruct.BootableImageVersion: // If old bootable version != given bootable version
		updateBootableImage = true
	}

	// Recreate DataDisk first if needed.
	if updateExternalDisk {
		logrus.Infof("Recreate data disk: %s", initOpts.ImagesStruct.DataDisk)
		err = system2.CreateAndResizeDisk(initOpts.ImagesStruct.DataDisk, strongunits.GiB(100)) //nolint:mnd
		if err != nil {
			return fmt.Errorf("failed to create/resize data disk: %w", err)
		}

		if oldMc != nil {
			// If old machine exists, update the DataDiskVersion field
			oldMc.DataDiskVersion = initOpts.ImageVerStruct.DataDiskVersion
			logrus.Infof("Update old machine configure DataDiskVersion field: %s", oldMc.ConfigPath.GetPath())
			if err = oldMc.Write(); err != nil {
				return fmt.Errorf("update machine configure error: %w", err)
			}
		}
	} else {
		logrus.Infof("Skip initialize data disk.")
	}

	if err = systemResourceCheck(cmd); err != nil {
		return err
	}

	// Update the machine configure  Resources field
	if oldMc != nil {
		err = updateMachineResource(oldMc)
		if err != nil {
			return err
		}
	}

	if !updateBootableImage {
		logrus.Infof("Skip initialize virtual machine with %s", initOpts.Name)
		return nil
	}

	for idx, vol := range initOpts.Volumes {
		initOpts.Volumes[idx] = os.ExpandEnv(vol)
	}

	logrus.Infof("Initialize virtual machine [ %s ] with bootable image: [ %s ]", initOpts.Name, initOpts.ImagesStruct.BootableImage)
	err = shim.Init(initOpts, provider)
	if err != nil {
		return fmt.Errorf("initialize virtual machine error: %w", err)
	}
	return nil
}

func systemResourceCheck(cmd *cobra.Command) error {
	if cmd.Flags().Changed("memory") {
		if err := system2.CheckMaxMemory(strongunits.MiB(initOpts.Memory)); err != nil {
			logrus.Errorf("Can not allocate the memory size %d", initOpts.Memory)
			return fmt.Errorf("can not allocate the memory size %d: %w", initOpts.Memory, err)
		}
	}

	// Krun limited max cpus core to 8
	if cmd.Flags().Changed(cmdflags.CpusFlag) {
		if initOpts.CPUS > cmdflags.KrunMaxCpus || initOpts.CPUS < 1 {
			return fmt.Errorf("can not allocate the CPU size %d", initOpts.CPUS)
		}
	}

	return nil
}

func updateMachineResource(mc *vmconfigs.MachineConfig) error {
	if mc != nil {
		mc.Resources.CPUs = initOpts.CPUS                                               // Update the CPUs
		mc.Resources.Memory = strongunits.MiB(initOpts.Memory)                          // Update the Memory
		mc.Mounts = shim.CmdLineVolumesToMounts(initOpts.Volumes, provider.MountType()) // Update the Volumes
		logrus.Infof("Update old machine CPU/Memory/Mounts configure: [ %s ]", mc.ConfigPath.GetPath())
		if err := mc.Write(); err != nil {
			return fmt.Errorf("update machine configure error: %w", err)
		}
	}

	return nil
}
