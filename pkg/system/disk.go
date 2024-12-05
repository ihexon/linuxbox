//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package system

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/containers/common/pkg/strongunits"
)

func CreateAndResizeDisk(diskPath string, newSize strongunits.GiB) error {
	if err := os.RemoveAll(diskPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to remove disk: %s, %w", diskPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(diskPath), 0755); err != nil {
		return fmt.Errorf("failed to create disk directory: %s, %w", diskPath, err)
	}

	file, err := os.Create(diskPath)
	if err != nil {
		return fmt.Errorf("failed to create disk: %s, %w", diskPath, err)
	}
	defer file.Close()
	if err = os.Truncate(diskPath, int64(newSize.ToBytes())); err != nil {
		return fmt.Errorf("failed to truncate disk: %w", err)
	}
	return nil
}
