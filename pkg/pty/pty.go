//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package pty

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/creack/pty"
)

const (
	rows = uint16(200)
	cols = uint16(200)
)

func RunInPty(c *exec.Cmd) (*os.File, error) {
	return pty.StartWithAttrs(c, &pty.Winsize{ //nolint:wrapcheck
		Cols: cols,
		Rows: rows,
		X:    rows,
		Y:    cols,
	}, &syscall.SysProcAttr{
		Setpgid: true,
	})
}
