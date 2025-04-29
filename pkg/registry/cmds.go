//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package registry

import (
	"os/exec"
)

var (
	cmds = make([]*exec.Cmd, 0)
)

func RegistryCmd(cmd *exec.Cmd) {
	cmds = append(cmds, cmd)
}

func GetCmds() []*exec.Cmd {
	return cmds
}
