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

	"bauklotze/pkg/machine/workspace"

	"github.com/sirupsen/logrus"

	"github.com/containers/common/pkg/strongunits"
)

type PathWrapper struct {
	path  string
	isDir bool
}

func NewFile(f string) *PathWrapper {
	fileWp := PathWrapper{
		path:  f,
		isDir: false,
	}
	return &fileWp
}

func NewDir(d string) *PathWrapper {
	dirWp := PathWrapper{
		path:  d,
		isDir: true,
	}
	return &dirWp
}

func (m *PathWrapper) GetPath() string {
	return m.path
}

func (m *PathWrapper) Delete(safe bool) error {
	workspaceDir := workspace.GetWorkspace()
	if workspaceDir == "" {
		return errors.New("workspace dir is empty")
	}

	if safe {
		if !strings.HasPrefix(m.path, workspaceDir) {
			return fmt.Errorf("path %q is not a workspace, refuse delete", m.path)
		}
	}

	logrus.Infof("delete file %s", m.path)
	if err := os.RemoveAll(m.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err //nolint:wrapcheck
	}

	return nil
}

// Read the contents of a given file and return in []bytes
func (m *PathWrapper) Read() ([]byte, error) {
	if m.isDir {
		return nil, fmt.Errorf("can not read content from directory %s", m.path)
	}
	return os.ReadFile(m.GetPath()) //nolint:wrapcheck
}

func (m *PathWrapper) IsExist() bool {
	_, err := os.Stat(m.path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

// DiscardBytesAtBegin discards the first n MB of a file
func (m *PathWrapper) DiscardBytesAtBegin(n strongunits.MiB) error {
	fileInfo, err := os.Stat(m.path)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	offset := int64(n.ToBytes())
	if fileInfo.Size() <= offset {
		return nil
	} else {
		file, _ := os.OpenFile(m.path, os.O_RDONLY, 0) // open the file in read-only mode
		defer file.Close()                             //nolint:errcheck

		_, _ = file.Seek(offset, io.SeekStart)
		tempFile, _ := os.CreateTemp("", "trimmed-ovm-log.txt")
		defer tempFile.Close() //nolint:errcheck

		_, _ = io.Copy(tempFile, file)
		_ = file.Close()
		_ = tempFile.Close()

		_ = os.Rename(tempFile.Name(), m.path)
	}
	return nil
}

func (m *PathWrapper) MakeBaseDir() error {
	err := os.MkdirAll(filepath.Dir(m.GetPath()), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create base dir: %w", err)
	}
	return nil
}

func (m *PathWrapper) AppendDir(additionalPath string) *PathWrapper {
	return NewDir(
		filepath.Join(m.path, additionalPath),
	)
}

func (m *PathWrapper) AppendFile(additionalPath string) *PathWrapper {
	return NewFile(filepath.Join(m.path, additionalPath))
}
