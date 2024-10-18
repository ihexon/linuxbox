package diskpull

import (
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/diskpull/internal/provider"
	"bauklotze/pkg/machine/diskpull/stdpull"
	"fmt"
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
		mydisk, err = stdpull.NewStdDiskPull(userInputPath, imagePath)
	}
	if err != nil {
		return err
	}
	return mydisk.Get()
}
