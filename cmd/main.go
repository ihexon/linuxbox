package main

import (
	_ "bauklotze/cmd/bauklotze/machine"
	"bauklotze/cmd/registry"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func flagErrorFuncfunc(c *cobra.Command, e error) error {
	e = fmt.Errorf("%w\nSee '%s --help'", e, c.CommandPath())
	return e
}

func main() {
	rootCmd = parseCommands()
	Execute()
	os.Exit(0)
}

func parseCommands() *cobra.Command {
	for _, c := range registry.Commands {
		addCommand(c)
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
	c.Command.SetHelpTemplate(helpTemplate)
	c.Command.SetUsageTemplate(usageTemplate)
	c.Command.DisableFlagsInUseLine = true
}
