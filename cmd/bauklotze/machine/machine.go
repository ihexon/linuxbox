package machine

import (
	cmdflags "bauklotze/cmd/bauklotze/flags"
	"bauklotze/cmd/bauklotze/validata"
	"bauklotze/cmd/registry"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var machineCmd = &cobra.Command{
	Use:   "machine",
	Short: "Manage a virtual machine",
	Long:  "Manage a virtual machine. Virtual machines are used to run OVM.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logrus.Infof("======== machineCmd PersistentPreRunE  ========")
		BAUKLOTZE_HOME := cmd.Flag(cmdflags.WorkspaceFlag).Value.String()
		logrus.Infof("set env %s: %s", cmdflags.BAUKLOTZE_HOME, BAUKLOTZE_HOME)
		_ = os.Setenv(cmdflags.BAUKLOTZE_HOME, BAUKLOTZE_HOME)
		return nil
	},
	//PersistentPostRunE: closeMachineEvents,
	RunE: validata.SubCommandExists,
}

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: machineCmd,
	})
}

//func closeMachineEvents(cmd *cobra.Command, _ []string) error {
//	logrus.Infof("machineCmd PersistentPostRunE")
//	return nil
//}
