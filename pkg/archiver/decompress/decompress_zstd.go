//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package decompress

import (
	"os"

	"bauklotze/pkg/machine/define"

	"github.com/DataDog/zstd"
)

func DecompressZstd(compressedFilePath *define.VMFile, decompressedFilePath *define.VMFile) error {
	var err error
	file, err := os.ReadFile(compressedFilePath.GetPath())
	if err != nil {
		return err
	}

	decompressData, err := zstd.Decompress(nil, file)
	if err != nil {
		return err
	}

	err = os.WriteFile(decompressedFilePath.GetPath(), decompressData, 0644)
	if err != nil {
		return err
	}

	return err
}
