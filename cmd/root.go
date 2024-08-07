package main

import (
	"bauklotze/cmd/bauklotze/validata"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

var (
	rootCmd = &cobra.Command{
		Use:                   filepath.Base(os.Args[0]) + " [options]",
		Long:                  "Manage pods, containers and images",
		SilenceUsage:          true,
		SilenceErrors:         true,
		TraverseChildren:      true,
		PersistentPreRunE:     persistentPreRunE,
		RunE:                  validata.SubCommandExists,
		PersistentPostRunE:    persistentPostRunE,
		DisableFlagsInUseLine: true,
	}

	requireCleanup = true
)

func Execute() {
	rootCmd.Execute()
}

func persistentPreRunE(cmd *cobra.Command, args []string) error {
	logrus.Debugf("Called %s.PersistentPreRunE(%s)", cmd.Name(), strings.Join(os.Args, " "))

	if cmd.Name() == "help" || cmd.Name() == "completion" || cmd.HasSubCommands() {
		requireCleanup = false
		return nil
	}

	return nil
}
func persistentPostRunE(cmd *cobra.Command, args []string) error {
	return nil
}
