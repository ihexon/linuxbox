//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package diskpull

import (
	"fmt"

	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/diskpull/stdpull"
)

// GetDisk For now we don't need dirs *define.MachineDirs,vmType define.VMType, name string
func GetDisk(userInputPath string, imagePath *define.VMFile) error {
	if userInputPath == "" {
		return fmt.Errorf("please provide a bootable image using --boot [IMAGE_PATH]")
	}

	mydisk, err := stdpull.NewStdDiskPull(userInputPath, imagePath)
	if err != nil {
		return fmt.Errorf("failed to create disk puller: %w", err)
	}

	return mydisk.Get() //nolint:wrapcheck
}
