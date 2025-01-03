//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package stdpull

import (
	"fmt"

	"bauklotze/pkg/decompress"
	"bauklotze/pkg/machine/define"

	"github.com/containers/storage/pkg/fileutils"
	"github.com/sirupsen/logrus"
)

type StdDiskPull struct {
	// all define.VMFile are not dir instead the full path contained file name
	inputPath *define.VMFile
	finalPath *define.VMFile
}

func NewStdDiskPull(inputPath string, finalpath *define.VMFile) (*StdDiskPull, error) {
	inputImage, err := define.NewMachineFile(inputPath, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create new machine file: %s, %w", inputPath, err)
	}
	return &StdDiskPull{inputPath: inputImage, finalPath: finalpath}, nil
}

// Get StdDiskPull: Get just decompress the `inputPath *define.VMFile` to `finalPath *define.VMFile`
// Nothing interesting at all
func (s *StdDiskPull) Get() error {
	if err := fileutils.Exists(s.inputPath.GetPath()); err != nil {
		return fmt.Errorf("could not find user input disk: %w", err)
	}
	logrus.Infof("Try to decompress %s to %s", s.inputPath.GetPath(), s.finalPath.GetPath())
	// Only support zstd compressed bootable.img
	err := decompress.Zstd(s.inputPath.GetPath(), s.finalPath.GetPath())
	if err != nil {
		errors := fmt.Errorf("could not decompress %s to %s, %w", s.inputPath.GetPath(), s.finalPath.GetPath(), err)
		return errors
	}
	return nil
}
