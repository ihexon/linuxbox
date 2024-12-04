//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin && !windows && !linux

package env

import "path/filepath"

// getTmpDir return ${BauklotzeHomePath}/tmp/
func getTmpDir() (string, error) {
	p, err := GetBauklotzeHomePath()
	if err != nil {
		return "", err
	}

	return filepath.Join(p, "tmp"), nil // ${BauklotzeHomePath}/tmp/
}

// getRuntimeDir: ${BauklotzeHomePath}/tmp/
func getRuntimeDir() (string, error) {
	return getTmpDir() // ${BauklotzeHomePath}/tmp/
}
