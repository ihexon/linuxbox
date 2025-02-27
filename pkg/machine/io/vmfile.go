//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package io

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	allFlag "bauklotze/pkg/machine/allflag"

	"github.com/sirupsen/logrus"

	"github.com/containers/common/pkg/strongunits"
)

type FileWrapper struct {
	Path string `json:"path,omitempty"`
}

// NewMachineFile is a constructor for FileWrapper
func NewMachineFile(f string) (*FileWrapper, error) {
	if len(f) < 1 {
		return nil, errors.New("invalid file path, must be at least 1 character")
	}

	mf := FileWrapper{Path: f}
	return &mf, nil
}

// GetPath returns the working path for a machinefile.  it returns
// the symlink unless one does not exist
func (m *FileWrapper) GetPath() string {
	return m.Path
}

// Delete dangerous removes a file from the filesystem
// if safety is true, it will only remove files in the workspace
func (m *FileWrapper) Delete(safety bool) error {
	if safety {
		workspace := allFlag.WorkSpace
		if workspace == "" {
			return fmt.Errorf("workspace is not set")
		}
		if !strings.HasPrefix(m.Path, workspace) {
			return fmt.Errorf("invicible path %s, can not delet non-workspaces file", m.Path)
		}
	}
	logrus.Warnf("Delete file %s", m.Path)
	if err := os.RemoveAll(m.Path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to remove %s : %w", m.Path, err)
	}

	return nil
}

// Read the contents of a given file and return in []bytes
func (m *FileWrapper) Read() ([]byte, error) {
	return os.ReadFile(m.GetPath()) //nolint:wrapcheck
}

func (m *FileWrapper) Exist() bool {
	_, err := os.Stat(m.Path)
	return err == nil
}

// DiscardBytesAtBegin discards the first n MB of a file
func (m *FileWrapper) DiscardBytesAtBegin(n strongunits.MiB) error {
	fileInfo, err := os.Stat(m.Path)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	offset := int64(n.ToBytes())
	if fileInfo.Size() <= offset {
		return nil
	} else {
		file, _ := os.OpenFile(m.Path, os.O_RDONLY, 0) // open the file in read-only mode
		defer file.Close()

		_, _ = file.Seek(offset, io.SeekStart)
		tempFile, _ := os.CreateTemp("", "trimmed-ovm-log.txt")
		defer tempFile.Close()

		_, _ = io.Copy(tempFile, file)
		file.Close()
		tempFile.Close()

		_ = os.Rename(tempFile.Name(), m.Path)
	}
	return nil
}

// AppendToNewVMFile takes a given path and appends it to the existing vmfile path.  The new
// FileWrapper is returned
func (m *FileWrapper) AppendToNewVMFile(additionalPath string) (*FileWrapper, error) {
	if additionalPath == "" {
		return nil, errors.New("invalid additional path")
	}
	return NewMachineFile(filepath.Join(m.Path, additionalPath))
}

func (m *FileWrapper) MakeBaseDir() error {
	err := os.MkdirAll(filepath.Dir(m.GetPath()), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create base dir: %w", err)
	}
	return nil
}
