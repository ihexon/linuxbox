//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build (darwin || linux) && (amd64 || arm64)

package terminal

// SetConsole for non-windows environments is a no-op.
func SetConsole() error {
	return nil
}
