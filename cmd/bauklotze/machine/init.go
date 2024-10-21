package machine

import (
	"bauklotze/cmd/registry"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/system"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/network"
	"bauklotze/pkg/notifyexit"
	system2 "bauklotze/pkg/system"
	"context"
	"errors"
	"fmt"
	"github.com/containers/common/pkg/strongunits"
	"github.com/containers/storage/pkg/regexp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"os"
	"path/filepath"
	"time"
)

var (
	NameRegex     = regexp.Delayed("^[a-zA-Z0-9][a-zA-Z0-9_.-]*$")
	RegexError    = fmt.Errorf("names must match [a-zA-Z0-9][a-zA-Z0-9_.-]*: %w", ErrInvalidArg)
	ErrInvalidArg = errors.New("invalid argument")
)

var (
	initCmd = &cobra.Command{
		Use:               "init [options] [NAME]",
		Short:             "initialize a virtual machine",
		Long:              "initialize a virtual machine",
		PersistentPreRunE: machinePreRunE,
		RunE:              initMachine,
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args:    cobra.MaximumNArgs(1), // max positional arguments
		Example: `machine init`,
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
	flags.StringVar(&initOpts.ImageVersion.BootableImageVersion, BootImageVersion, cfg.ContainersConfDefaultsRO.Machine.Image, "Boot version field")

	DataImageVersion := dataVersion
	flags.StringVar(&initOpts.ImageVersion.DataDiskVersion, DataImageVersion, "", "Data version field")

	sendEventToEndpoint := reportUrlFlag
	flags.StringVar(&initOpts.SendEvt, sendEventToEndpoint, "", "send events to somewhere, only support unix:///....")

	ppidFlagName := ppid
	flags.Int32Var(&initOpts.PPID, ppidFlagName, -1, "Parent process id, if not given, the ppid is the current process's ppid")
}

func initMachine(cmd *cobra.Command, args []string) error {
	network.NewReporter(initOpts.SendEvt)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)
	// If not specified PPID, use the current process id as the parent process id
	if initOpts.PPID == -1 {
		mypid := os.Getpid()
		initOpts.PPID = int32(mypid)
	}

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return context.Cause(ctx)
			default:
			}
			if isRunning, err := system.IsProcesSAlive([]int32{initOpts.PPID}); !isRunning {
				return err
			}
			time.Sleep(300 * time.Millisecond)
		}
	})

	go func() {
		if err := g.Wait(); err != nil && !(errors.Is(err, context.Canceled)) {
			logrus.Errorf("%s\n", err.Error())
			network.Reporter.SendEventToOvmJs("error", err.Error())
			registry.SetExitCode(1)
			notifyexit.NotifyExit(registry.GetExitCode())
		}
	}()

	// Initialize the network reporter

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

	dataDir, err := env.DataDirPrefix() // ${BauklotzeHomePath}/data
	if err != nil {
		return fmt.Errorf("can not get Data dir %v", err)
	}

	dataDisk := filepath.Join(dataDir, "external_disk", initOpts.Name, "data.raw") // ${BauklotzeHomePath}/data/{MachineName}/data.raw
	initOpts.Images.DataDisk = dataDisk

	overlayDisk := filepath.Join(dataDir, "external_disk", initOpts.Name, "overlay.raw") // ${BauklotzeHomePath}/data/{MachineName}/overlay.raw
	initOpts.Images.OverlayImage = overlayDisk

	var (
		updateBootableImage bool = true
		updateExternalDisk  bool = true
		updateOverlayDisk   bool = true
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

	// Notice the  updateOverlayDisk always be true for now !
	if updateOverlayDisk {
		logrus.Infof("Recreate overlay disk: %s", initOpts.Images.OverlayImage)
		err = system2.CreateAndResizeDisk(initOpts.Images.OverlayImage, strongunits.GiB(100))
		if err != nil {
			return err
		}
	}

	if !updateBootableImage {
		msg := "skip initialize virtual machine"
		logrus.Errorf(msg)
		return fmt.Errorf(msg)
	}

	for idx, vol := range initOpts.Volumes {
		initOpts.Volumes[idx] = os.ExpandEnv(vol)
	}

	// The allocate virtual memory can not bigger than physic virtual memory
	if cmd.Flags().Changed("memory") {
		if err := system2.CheckMaxMemory(strongunits.MiB(initOpts.Memory)); err != nil {
			logrus.Infof("Can not allocate the memory size %s", initOpts.Memory)
			return err
		}
	}

	err = shim.Init(initOpts, provider)
	if err != nil {
		network.Reporter.SendEventToOvmJs("error", err.Error())
		return err
	}
	return nil
}
