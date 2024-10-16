package system

import (
	"errors"
	"github.com/containers/common/pkg/strongunits"
	"os"
	"path/filepath"
)

func CreateAndResizeDisk(diskPath string, newSize strongunits.GiB) error {
	if err := os.RemoveAll(diskPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(diskPath), 0755); err != nil {
		return err
	}
	
	file, err := os.Create(diskPath)
	if err != nil {
		return err
	}
	defer file.Close()
	if err = os.Truncate(diskPath, int64(newSize.ToBytes())); err != nil {
		return err
	}
	return nil
}
