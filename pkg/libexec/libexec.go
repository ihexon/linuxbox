//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package libexec

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	libexecDir string
)

func Setup(executablePath string) error {
	libexecDir = filepath.Join(filepath.Dir(filepath.Dir(executablePath)), "libexec")
	if _, err := os.Stat(libexecDir); err != nil {
		return fmt.Errorf("failed to find libexec directory: %w", err)
	}

	return nil
}

func FindBinary(name string) (string, error) {
	p := filepath.Join(libexecDir, name)
	if _, err := os.Stat(p); err != nil {
		return "", fmt.Errorf("failed to find %q in libexec directory: %w", name, err)
	}
	return filepath.Join(libexecDir, name), nil
}

// GetDYLDLibraryPath contains the dyld files
func GetDYLDLibraryPath() string {
	return libexecDir
}
