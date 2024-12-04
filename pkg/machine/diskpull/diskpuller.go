//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package diskpull

import (
	"fmt"

	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/diskpull/internal/provider"
	"bauklotze/pkg/machine/diskpull/stdpull"
)

// GetDisk For now we don't need dirs *define.MachineDirs,vmType define.VMType, name string
func GetDisk(userInputPath string, imagePath *define.VMFile) error {
	var (
		err    error
		mydisk provider.Disker
	)
	switch {
	case userInputPath == "":
		return fmt.Errorf("please provide a bootable image using --boot [IMAGE_PATH]")
	default:
		zstdFile := &userInputPath
		extractFile := &imagePath
		mydisk, err = stdpull.NewStdDiskPull(*zstdFile, *extractFile)
	}
	if err != nil {
		return err
	}
	return mydisk.Get()
}
