//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package env

import (
	"fmt"
	"path/filepath"
)

func getRuntimeDir() (string, error) {
	p, err := GetBauklotzeHomePath()
	if err != nil {
		return "", fmt.Errorf("unable to get home path: %w", err)
	}

	return filepath.Join(p, "tmp"), nil
}
