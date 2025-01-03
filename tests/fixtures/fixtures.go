//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package fixtures

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var dir string
var once sync.Once

func GetTestFixtures(name string) string {
	once.Do(func() {
		cwd, err := os.Getwd()
		if err != nil {
			panic(fmt.Errorf("failed to get current working directory: %w", err))
		}
		dir = getTestFixtures(cwd)
	})

	return filepath.Join(dir, name)
}

func getTestFixtures(dir string) string {
	if dir == "/" || filepath.VolumeName(dir) == dir {
		panic(fmt.Errorf("could not find project root (no go.mod found)"))
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		panic(fmt.Errorf("failed to read current working directory: %w", err))
	}

	for _, entry := range entries {
		if entry.Name() == "go.mod" {
			return filepath.Join(dir, "tests", "fixtures")
		}
	}

	return getTestFixtures(filepath.Dir(dir))
}
