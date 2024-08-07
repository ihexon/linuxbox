package main

import (
	_ "bauklotze/cmd/bauklotze/machine"
	"bauklotze/cmd/registry"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	rootCmd = parseCommands()
	Execute()
	os.Exit(0)
}

func parseCommands() *cobra.Command {
	for _, c := range registry.Commands {
		addCommand(c)
	}
	return rootCmd
}

func addCommand(c registry.CliCommand) {
	parent := rootCmd
	if c.Parent != nil {
		parent = c.Parent
	}
	parent.AddCommand(c.Command)
}
