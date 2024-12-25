//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package vmconfigs

import "strings"

func pathsFromVolume(volume string) []string {
	return strings.SplitN(volume, ":", 3) //nolint:mnd
}

func extractTargetPath(paths []string) string {
	if len(paths) > 1 {
		return paths[1] + "/" // Add trailing slash to target path
	}
	return paths[0] + "/" // Add trailing slash to target path
}
