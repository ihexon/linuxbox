package machineDefine

import (
	"errors"
	"os"
	"path/filepath"
)

type VMFile struct {
	Path string
}

func (m *VMFile) GetPath() string {
	return m.Path
}

func (m *VMFile) GetAbsPath() (string, error) {
	abs, err := filepath.Abs(m.Path)
	if err != nil {
		return "", err
	}
	return abs, nil
}

func (m *VMFile) Abs() error {
	path, err := m.GetAbsPath()
	if err != nil {
		return err
	}
	m.Path = path
	return nil
}

func (m *VMFile) CreatePath() error {
	return os.MkdirAll(m.Path, 0755)
}

func (m *VMFile) CreateFile() error {
	f, err := os.OpenFile(m.Path, os.O_CREATE, 0o755)
	defer f.Close()
	if err != nil {
		return err
	}
	return nil
}

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
