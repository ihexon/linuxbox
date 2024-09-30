package machine

import (
	"bauklotze/cmd/bauklotze/validata"
	"bauklotze/cmd/registry"
	"github.com/spf13/cobra"
)

var (
	// Pull in configured json library
	json = registry.JSONLibrary()

	// Command: podman _system_
	systemCmd = &cobra.Command{
		Use:   "system",
		Short: "Manage podman",
		Long:  "Manage podman",
		RunE:  validata.SubCommandExists,
	}
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: systemCmd,
	})
}
