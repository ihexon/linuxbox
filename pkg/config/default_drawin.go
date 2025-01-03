//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package config

func getDefaultMachineVolumes() []string {
	// Empty mount point
	return []string{}
}
