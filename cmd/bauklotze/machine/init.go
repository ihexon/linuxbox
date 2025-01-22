//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package machine

import (
	"bauklotze/cmd/registry"
	"bauklotze/pkg/decompress"
	allFlag "bauklotze/pkg/machine/allflag"
	"bauklotze/pkg/machine/defconfig"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/events"
	"bauklotze/pkg/machine/helper"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/vmconfig"
	"bauklotze/pkg/machine/volumes"
	"bauklotze/pkg/system"
	"errors"
	"fmt"
	"os"

	"github.com/containers/common/pkg/strongunits"

	"github.com/sirupsen/logrus"

	"github.com/containers/storage/pkg/regexp"
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
		PersistentPreRunE: registry.PersistentPreRunE,
		PreRunE:           registry.PreRunE,
		RunE:              initMachine,
		Args:              cobra.MaximumNArgs(1), // max positional arguments
		Example:           `machine init default`,
	}
)

var cfg *defconfig.DefaultConfig

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: initCmd,
		Parent:  machineCmd,
	})

	cfg = defconfig.VMConfig()
	lFlags := initCmd.Flags()

	ppidFlagName := registry.PpidFlag
	defaultPPID, _ := system.GetPPID(int32(os.Getpid()))

	lFlags.Int32Var(&allFlag.PPID, ppidFlagName, defaultPPID,
		"Parent process id, if not given, the ppid is the current process's ppid")

	cpusFlagName := registry.CpusFlag
	lFlags.Uint64Var(
		&allFlag.CPUS,
		cpusFlagName, cfg.CPUs,
		"Number of CPUs",
	)

	memoryFlagName := registry.MemoryFlag
	lFlags.Uint64VarP(
		&allFlag.Memory,
		memoryFlagName, "m", cfg.Memory,
		"Memory in MiB",
	)

	VolumeFlagName := registry.VolumeFlag
	lFlags.StringArrayVarP(&allFlag.Volumes, VolumeFlagName, "v", []string{}, "Volumes to mount, source:target")

	BootImageName := registry.BootImageFlag
	lFlags.StringVar(&allFlag.BootableImage, BootImageName, cfg.Image, "Bootable image for machine")
	_ = initCmd.MarkFlagRequired(BootImageName)

	BootImageVersion := registry.BootVersionFlag
	lFlags.StringVar(&allFlag.BootableImageVersion, BootImageVersion, cfg.Image, "Boot version field")
	_ = initCmd.MarkFlagRequired(BootImageVersion)

	DataImageVersion := registry.DataVersionFlag
	lFlags.StringVar(&allFlag.DataDiskVersion, DataImageVersion, "", "Data version field")
	_ = initCmd.MarkFlagRequired(DataImageVersion)
}

func initMachine(cmd *cobra.Command, args []string) error {
	logrus.Infof("===================INIT===================")
	vmp, err := registry.GetProvider()
	if err != nil {
		logrus.Errorf("failed to get current hypervisor provider: %v", err)
	}

	// Inject default mnt point from cfg into allflag.Volumes
	allFlag.Volumes = append(allFlag.Volumes, cfg.Volumes.Get()...)

	mc, err := shim.GetVMConf(allFlag.VMName, []vmconfig.VMProvider{vmp})
	// err != nil means the machine config not find and need to initialize the VM
	if err != nil {
		logrus.Warnf("Get machine configure error: %v, try to initialize the VM", err)
		events.NotifyInit(events.InitNewMachine)
		if err = initializeNewVM(vmp); err != nil {
			return fmt.Errorf("initialize virtual machine error: %w", err)
		}
	} else {
		mc.VMProvider = vmp
		// Machine already exists, update the machine configure
		logrus.Infof("Machine %q already exists, update the machine configure", allFlag.VMName)
		events.NotifyInit(events.InitUpdateConfig)
		if err = updateExistedMachine(mc); err != nil {
			return fmt.Errorf("update machine configure error: %w", err)
		}
	}
	return nil
}

func updateExistedMachine(mc *vmconfig.MachineConfig) error {
	mc.Resources.CPUs = allFlag.CPUS
	mc.Resources.Memory = strongunits.MiB(allFlag.Memory)
	mc.Mounts = volumes.CmdLineVolumesToMounts(allFlag.Volumes)
	if err := updateImage(mc); err != nil {
		return fmt.Errorf("update image error: %w", err)
	}
	if err := mc.Write(); err != nil {
		return fmt.Errorf("write machine configure error: %w", err)
	}
	return nil
}

func updateImage(mc *vmconfig.MachineConfig) error {
	if mc.Bootable.Version != allFlag.BootableImageVersion {
		logrus.Warnf("Bootable image version is not match, try to update boot image")
		if err := updateBootImage(mc); err != nil {
			return fmt.Errorf("update boot image error: %w", err)
		}
	}
	if mc.DataDisk.Version != allFlag.DataDiskVersion {
		logrus.Warnf("Data image version is not match, try to update data image")
		if err := updateDataImage(mc); err != nil {
			return fmt.Errorf("update data image error: %w", err)
		}
	}
	mc.Bootable.Version = allFlag.BootableImageVersion
	mc.DataDisk.Version = allFlag.DataDiskVersion
	return nil
}

func updateDataImage(mc *vmconfig.MachineConfig) error {
	if err := helper.CreateAndResizeDisk(mc.DataDisk.Image, define.DefaultDataImageSizeGB); err != nil {
		return fmt.Errorf("failed to create and resize disk: %w", err)
	}
	return nil
}

func updateBootImage(mc *vmconfig.MachineConfig) error {
	destDir := mc.Bootable.Image.GetPath()
	logrus.Infof("Try to decompress %s to %s", allFlag.BootableImage, destDir)
	if err := decompress.Zstd(allFlag.BootableImage, destDir); err != nil {
		return fmt.Errorf("could not decompress %s to %s, %w", allFlag.BootableImage, destDir, err)
	}
	return nil
}
func initializeNewVM(mp vmconfig.VMProvider) error {
	if err := shim.Init(mp); err != nil {
		return fmt.Errorf("initialize virtual machine error: %w", err)
	}
	return nil
}
