package diskpull

import (
	"bauklotze/cmd/registry"
	"bauklotze/pkg/machine/diskpull/internal/provider"
	"bauklotze/pkg/machine/diskpull/stdpull"
	"bauklotze/pkg/machine/machineDefine"
	"fmt"
	"strings"
)

// GetDisk For now we don't need dirs *machineDefine.MachineDirs,vmType machineDefine.VMType, name string
// But I prefer the function signature same as podman original, so the VMProvider same as podman.
// We can just import any libraries from containers/* because we have the same function signature
func GetDisk(userInputPath string, dirs *machineDefine.MachineDirs, imagePath *machineDefine.VMFile, vmType machineDefine.VMType, name string) error {
	var (
		err    error
		mydisk provider.Disker
	)

	switch {
	case userInputPath == "":
		return fmt.Errorf("Please --image [IMAGE_PATH]")
	case strings.HasPrefix(userInputPath, "http"):
		return fmt.Errorf("Do not support download image from http(s)://")
	case strings.HasPrefix(userInputPath, "docker://"):
		return fmt.Errorf("Do not support download image from docker://")
	default:
		mydisk, err = stdpull.NewStdDiskPull(userInputPath, imagePath)
	}
	if err != nil {
		registry.SetExitCode(1)
		return err
	}
	return mydisk.Get()
}
