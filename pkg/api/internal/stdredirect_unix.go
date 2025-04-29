//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build (darwin || linux) && (arm64 || amd64)

package internal

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func RedirectStdin() error {
	devNullfile, err := os.Open(os.DevNull)
	if err != nil {
		return fmt.Errorf("failed to open /dev/null: %w", err)
	}
	defer devNullfile.Close() //nolint:errcheck

	if err := unix.Dup2(int(devNullfile.Fd()), int(os.Stdin.Fd())); err != nil {
		return fmt.Errorf("failed /dev/null to redirect stdin: %w", err)
	}
	return nil
}
