package machineDefine

import (
	"errors"
	"os"
	"path/filepath"
)

func (m *VMFile) GetPath() string {
	return m.Path
}

// Delete removes the machinefile symlink (if it exists) and
// the actual path
func (m *VMFile) Delete() error {
	if err := os.Remove(m.Path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (m *VMFile) Read() ([]byte, error) {
	return os.ReadFile(m.GetPath())
}

// NewMachineFile is a constructor for VMFile
func NewMachineFile(path string) (*VMFile, error) {
	if len(path) < 1 {
		return nil, errors.New("invalid machine file path")
	}
	mf := VMFile{Path: path}
	return &mf, nil
}

func (m *VMFile) AppendToNewVMFile(additionalPath string) (*VMFile, error) {
	return NewMachineFile(filepath.Join(m.Path, additionalPath))
}
