//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package main

import (
	_ "bauklotze/cmd/bauklotze/machine"
	"bauklotze/cmd/registry"
	allFlag "bauklotze/pkg/machine/allflag"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/events"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:                   filepath.Base(os.Args[0]) + " [options]",
		Long:                  "Manage your bugbox",
		SilenceUsage:          true,
		SilenceErrors:         true,
		TraverseChildren:      true,
		DisableFlagsInUseLine: true,
	}
)

func init() {
	cobra.EnableTraverseRunHooks = false
	cobra.OnInitialize(
		loggingHook,
	)

	rootCmd.SetUsageTemplate(define.UsageTemplate)
	pFlags := rootCmd.PersistentFlags()

	outFlagName := registry.LogOutFlag
	pFlags.StringVar(&allFlag.LogOut, outFlagName, registry.FileBased, "If set --log-out console, send output to terminal, if set --log-out file, send output to ${workspace}/logs/ovm.log")

	workspace := registry.WorkspaceFlag
	pFlags.StringVar(&allFlag.WorkSpace, workspace, "", "Bauklotze's HOME directory, this workspace is mandatory required")
	_ = rootCmd.MarkPersistentFlagRequired(workspace)

	ReportURLFlag := registry.ReportURLFlag
	pFlags.StringVar(&allFlag.ReportURL, ReportURLFlag, "", "Report events to the url")
}

func main() {
	rootCmd = parseCommands()
	RootCmdExecute()
}

func flagErrorFunc(c *cobra.Command, e error) error {
	e = fmt.Errorf("%w\nSee '%s --help'", e, c.CommandPath())
	return e
}

func parseCommands() *cobra.Command {
	for _, c := range registry.Commands {
		addCommand(c)
	}

	rootCmd.SetFlagErrorFunc(flagErrorFunc)
	return rootCmd
}

func addCommand(c registry.CliCommand) {
	parent := rootCmd
	if c.Parent != nil {
		parent = c.Parent
	}
	parent.AddCommand(c.Command)
	c.Command.SetFlagErrorFunc(flagErrorFunc)
	c.Command.SetHelpTemplate(define.HelpTemplate)
	c.Command.SetUsageTemplate(define.UsageTemplate)
	c.Command.DisableFlagsInUseLine = true
}

func RootCmdExecute() {
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		logrus.Errorf("Exit duto error: %v", err)
		if errors.Is(err, define.ErrVMAlreadyRunning) {
			events.NotifyError(err)
		}
		registry.NotifyAndExit(1)
	} else {
		registry.NotifyAndExit(0)
	}
}

func loggingHook() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		ForceColors:     true,
		DisableColors:   false,
		TimestampFormat: "2006-01-02 15:04:05.000",
	})
	logrus.SetOutput(os.Stderr)
}
