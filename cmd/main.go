package main

import (
	_ "bauklotze/cmd/bauklotze/machine"
	"bauklotze/cmd/registry"
	"bauklotze/pkg/terminal"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func flagErrorFuncfunc(c *cobra.Command, e error) error {
	e = fmt.Errorf("%w\nSee '%s --help'", e, c.CommandPath())
	return e
}

func main() {
	rootCmd = parseCommands()
	RootCmdExecute()
	os.Exit(0)
}

func parseCommands() *cobra.Command {
	for _, c := range registry.Commands {
		addCommand(c)
	}

	if err := terminal.SetConsole(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	rootCmd.SetFlagErrorFunc(flagErrorFuncfunc)
	return rootCmd
}

func addCommand(c registry.CliCommand) {
	parent := rootCmd
	if c.Parent != nil {
		parent = c.Parent
	}
	parent.AddCommand(c.Command)
	c.Command.SetFlagErrorFunc(flagErrorFuncfunc)
	c.Command.SetHelpTemplate(helpTemplate)
	c.Command.SetUsageTemplate(usageTemplate)
	c.Command.DisableFlagsInUseLine = true
}
