//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

// Package workspace should not depend on any other packages.
package workspace

var (
	workspace string
)

func SetWorkspace(w string) {
	workspace = w
}

// GetWorkspace returns the current workspace directory path.
// This is the canonical way to access the workspace location.
func GetWorkspace() string {
	return workspace
}
