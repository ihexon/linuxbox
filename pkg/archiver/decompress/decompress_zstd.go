//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package decompress

import (
	"fmt"
	"os"

	"bauklotze/pkg/machine/define"

	"github.com/DataDog/zstd"
)

func Zstd(compressedFilePath *define.VMFile, decompressedFilePath *define.VMFile) error {
	file, err := os.ReadFile(compressedFilePath.GetPath())
	if err != nil {
		return fmt.Errorf("failed to read compressed file: %w", err)
	}

	decompressData, err := zstd.Decompress(nil, file)
	if err != nil {
		return fmt.Errorf("failed to decompress file: %w", err)
	}

	if err = os.WriteFile(decompressedFilePath.GetPath(), decompressData, 0644); err != nil {
		return fmt.Errorf("failed to write decompressed file: %w", err)
	}

	return nil
}
