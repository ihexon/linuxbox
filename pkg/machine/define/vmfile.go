//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package define

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type VMFile struct {
	// Path is the fully qualified path to a file
	Path string `json:"Path"`
	// Symlink is a shortened version of Path by using
	// a symlink
	Symlink *string `json:"symlink,omitempty"`
}

// GetPath returns the working path for a machinefile.  it returns
// the symlink unless one does not exist
func (m *VMFile) GetPath() string {
	if m.Symlink == nil {
		return m.Path
	}
	return *m.Symlink
}

// Delete removes the machinefile symlink (if it exists) and
// the actual path
func (m *VMFile) Delete() error {
	if m.Symlink != nil {
		if err := os.RemoveAll(*m.Symlink); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("unable to remove symlink %q", *m.Symlink)
		}
	}
	if err := os.RemoveAll(m.Path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to remove symlink: %w", err)
	}
	return nil
}

// Read the contents of a given file and return in []bytes
func (m *VMFile) Read() ([]byte, error) {
	return os.ReadFile(m.GetPath()) //nolint:wrapcheck
}

// NewMachineFile is a constructor for VMFile
func NewMachineFile(path string, symlink *string) (*VMFile, error) {
	if len(path) < 1 {
		return nil, errors.New("invalid machine file path")
	}
	if symlink != nil && len(*symlink) < 1 {
		return nil, errors.New("invalid symlink path")
	}
	mf := VMFile{Path: path}
	logrus.Debugf("socket length for %s is %d", path, len(path))
	return &mf, nil
}

// AppendToNewVMFile takes a given path and appends it to the existing vmfile path.  The new
// VMFile is returned
func (m *VMFile) AppendToNewVMFile(additionalPath string, symlink *string) (*VMFile, error) {
	return NewMachineFile(filepath.Join(m.Path, additionalPath), symlink)
}
