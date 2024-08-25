package diskpull

import (
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/shim/stdpull"
	"bauklotze/pkg/ovmdisk"
	"fmt"
	"strings"
)

// GetDisk For now we don't need dirs *define.MachineDirs,vmType define.VMType, name string
// But I prefer the function signature same as podman original, so the VMProvider same as podman.
// We can just import any libraries from containers/* because we have the same function signature :)
func GetDisk(userInputPath string, dirs *define.MachineDirs, imagePath *define.VMFile, vmType define.VMType, name string) error {
	var (
		err    error
		mydisk ovmdisk.Disker
	)

	switch {
	case userInputPath == "":
		fmt.Errorf("Please --image [IMAGE_PATH]")
	case strings.HasPrefix(userInputPath, "http"):
		fmt.Errorf("Do not support download image from http(s)://")
	case strings.HasPrefix(userInputPath, "docker://"):
		fmt.Errorf("Do not support download image from docker://")
	default:
		mydisk, err = stdpull.NewStdDiskPull(userInputPath, imagePath)
	}
	if err != nil {
		return err
	}
	return mydisk.Get()
}
