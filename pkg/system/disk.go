package system

import (
	"github.com/containers/common/pkg/strongunits"
	"os"
)

func CreateAndResizeDisk(diskPath string, newSize strongunits.GiB) error {
	_ = os.RemoveAll(diskPath)
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
