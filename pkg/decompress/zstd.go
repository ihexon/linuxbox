//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package decompress

import (
	"fmt"
	"os"

	"github.com/DataDog/zstd"
)

func Zstd(src, target string) error {
	file, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read compressed file: %w", err)
	}

	data, err := zstd.Decompress(nil, file)
	if err != nil {
		return fmt.Errorf("failed to decompress file: %w", err)
	}

	if err = os.WriteFile(target, data, 0644); err != nil {
		return fmt.Errorf("failed to write decompressed file: %w", err)
	}

	return nil
}
