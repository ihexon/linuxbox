package machine

import (
	"bauklotze/cmd/registry"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/system"
	"errors"
	"fmt"
	strongunits "github.com/containers/common/pkg/strongunits"
	"github.com/containers/storage/pkg/regexp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	NameRegex     = regexp.Delayed("^[a-zA-Z0-9][a-zA-Z0-9_.-]*$")
	RegexError    = fmt.Errorf("names must match [a-zA-Z0-9][a-zA-Z0-9_.-]*: %w", ErrInvalidArg) // nolint:revive // This lint is new and we do not want to break the API.
	ErrInvalidArg = errors.New("invalid argument")
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
	}
	initOpts = define.InitOptions{
		Username: "root",
	}
	defaultMachineName = define.DefaultMachineName
	now                bool
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: initCmd,
		Parent:  machineCmd,
	})

	// Calculate the default configuration
	// CPU,MEMORY,VOLUME,etc..
	// OvmInitConfig() 配置虚拟机的内存/CPU/磁盘大小/外部挂载节点，这些配置将被写入 machine 的 json 文件做到持久化
	cfg := registry.OvmInitConfig()

	flags := initCmd.Flags()
	flags.BoolVar(&now,
		"startnow",
		false,
		"Start machine now",
	)

	cpusFlagName := "cpus"
	flags.Uint64Var(
		&initOpts.CPUS,
		cpusFlagName, cfg.ContainersConfDefaultsRO.Machine.CPUs,
		"Number of CPUs",
	)
	//_ = initCmd.RegisterFlagCompletionFunc(cpusFlagName, completion.AutocompleteNone)

	memoryFlagName := "memory"
	flags.Uint64VarP(
		&initOpts.Memory,
		memoryFlagName, "m", cfg.ContainersConfDefaultsRO.Machine.Memory,
		"Memory in MiB",
	)

	VolumeFlagName := "volume"
	flags.StringArrayVarP(&initOpts.Volumes, VolumeFlagName, "v", cfg.ContainersConfDefaultsRO.Machine.Volumes.Get(), "Volumes to mount, source:target")

	ImageFlagName := "image"
	flags.StringVar(&initOpts.Image, ImageFlagName, cfg.ContainersConfDefaultsRO.Machine.Image, "Bootable image for machine")
	flags.MarkHidden(ImageFlagName)

	ExternalImageFlagName := "external-disk"
	flags.StringVar(&initOpts.ExternImage, ExternalImageFlagName, "", "External image for machine")

	twinPid := "twinpid"
	flags.IntVar(&initOpts.TwinPid, twinPid, -1, "self killing when [twin pid] exit")
	flags.MarkHidden(twinPid)

	imageVersion := "image-version"
	flags.StringVar(&initOpts.ImageVersion, imageVersion, "always-update", "Special bootable image version")
	flags.MarkHidden(imageVersion)

	sendEventToEndpoint := "evtsock"
	flags.StringVar(&initOpts.SendEvt, sendEventToEndpoint, "", "send events to somewhere")
	flags.MarkHidden(sendEventToEndpoint)
}

func initMachine(cmd *cobra.Command, args []string) error {

	if now {
		// Pass TwinPid/evtsock to startOpts
		startOpts.TwinPid = initOpts.TwinPid
		startOpts.SendEvt = initOpts.SendEvt
	}

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
		if err := system.CheckMaxMemory(strongunits.MiB(initOpts.Memory)); err != nil {
			return err
		}
	}

	err = shim.Init(initOpts, provider)
	if err != nil {
		return err
	}

	if now {
		logrus.Infof("starting machine now with %s", args)
		return start(nil, args)
	} else {
		fmt.Printf("To start your machine run:\n\n\tbauklotze machine start%s\n\n", initOpts.Name)
	}

	return nil
}
