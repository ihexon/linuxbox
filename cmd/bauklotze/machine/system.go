//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package machine

import (
	"bauklotze/cmd/bauklotze/validata"
	"bauklotze/cmd/registry"

	"github.com/spf13/cobra"
)

var (
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
