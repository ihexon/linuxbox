//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package defconfig

func getDefaultMachineVolumes() []string {
	return []string{
		"/tmp/initfs:/tmp/initfs",
	}
}
