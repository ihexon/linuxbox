package machine

import (
	"bauklotze/cmd/registry"
	"bauklotze/pkg/completion"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/regexp"
	strongunits "bauklotze/pkg/storage"
	"errors"
	"fmt"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	NameRegex     = regexp.Delayed("^[a-zA-Z0-9][a-zA-Z0-9_.-]*$")
	RegexError    = fmt.Errorf("names must match [a-zA-Z0-9][a-zA-Z0-9_.-]*: %w", ErrInvalidArg) // nolint:revive // This lint is new and we do not want to break the API.
	ErrInvalidArg = errors.New("invalid argument")
	NotHexRegex   = regexp.Delayed(`[^0-9a-fA-F]`)
)

const maxMachineNameSize = 30

var (
	initCmd = &cobra.Command{
		Use:               "init [options] [NAME]",
		Short:             "Reset and initialize a virtual machine",
		Long:              "Reset and initialize a virtual machine",
		PersistentPreRunE: machinePreRunE,
		RunE:              initMachine,
		Args:              cobra.MaximumNArgs(1), // max positional arguments
		Example:           `machine init`,
		ValidArgsFunction: completion.AutocompleteNone,
	}
	initOpts           = define.InitOptions{}
	defaultMachineName = define.DefaultMachineName
	now                bool
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
	// Calculate the default configuration
	// CPU,MEMORY,VOLUME,etc..
	// OvmInitConfig() 配置虚拟机的内存/CPU/磁盘大小，这些配置将被写入 machine 的 json 文件做到持久化
	cfg := registry.OvmInitConfig()

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

	rootfulFlagName := "rootful"
	flags.BoolVar(&initOpts.Rootful, rootfulFlagName, true, "Whether this machine should prefer rootful container execution")

	twinPid := "twinpid"
	flags.IntVar(&initOpts.TwinPid, twinPid, -1, "self killing when [twin pid] exit")
	flags.MarkHidden(twinPid)

	imageVersion := "image-version"
	flags.StringVar(&initOpts.ImageVersion, imageVersion, "always-update", "Special bootable image version")
	flags.MarkHidden(twinPid)

	sendEventToEndpoint := "evtsock"
	flags.StringVar(&initOpts.SendEvt, sendEventToEndpoint, "", "send events to somewhere")
	flags.MarkHidden(sendEventToEndpoint)

	flags.BoolVar(&now, "now", false, "Start machine now")
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

	oldmc, exists, err := shim.VMExists(initOpts.Name, []vmconfigs.VMProvider{provider})
	if err != nil {
		return err
	}

	switch {
	case exists == true && oldmc != nil && oldmc.ImageVersion != initOpts.ImageVersion:
		logrus.Infof("%s: %s", initOpts.Name, define.ErrVMAlreadyExists)
		logrus.Infof("New image-version:%s, old image-version: %s, Force Initialize....", initOpts.ImageVersion, oldmc.ImageVersion)
		break
	case initOpts.ImageVersion == "always-update":
		break
	case exists == true && oldmc != nil:
		logrus.Infof("%s: %s, skip initialize !", initOpts.Name, define.ErrVMAlreadyExists)
		if now {
			return start(cmd, args)
		}
		return fmt.Errorf("%s: %w", initOpts.Name, define.ErrVMAlreadyExists)
	case oldmc == nil:

	case oldmc.ImageVersion != initOpts.ImageVersion:
	default:
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
	//NewMachineEvent(events.Init, events.Event{Name: initOpts.Name})
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
