//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package machine

import (
	"os"

	cmdflags "bauklotze/cmd/bauklotze/flags"
	"bauklotze/cmd/bauklotze/validata"
	"bauklotze/cmd/registry"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var machineCmd = &cobra.Command{
	Use:   "machine",
	Short: "Manage a virtual machine",
	Long:  "Manage a virtual machine. Virtual machines are used to run OVM.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		v := cmd.Flag(cmdflags.WorkspaceFlag).Value.String()
		logrus.Infof("Set env %s: %s", cmdflags.BauklotzeHome, v)
		_ = os.Setenv(cmdflags.BauklotzeHome, v)
		return nil
	},
	RunE: validata.SubCommandExists,
}

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: machineCmd,
	})
}
