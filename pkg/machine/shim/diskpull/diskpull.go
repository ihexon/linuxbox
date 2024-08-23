package diskpull

import (
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/shim/stdpull"
	"bauklotze/pkg/ovmdisk"
	"strings"
)

// GetDisk For now we dont need dirs *define.MachineDirs,vmType define.VMType, name string
// But I prefer the function signature same as podman original, so the VMProvider same as podman.
// We can just import any libraries from containers/* because we have the same function signature :)
func GetDisk(userInputPath string, dirs *define.MachineDirs, imagePath *define.VMFile, vmType define.VMType, name string) error {
	var (
		err    error
		mydisk ovmdisk.Disker
	)

	if userInputPath == "" || strings.HasPrefix(userInputPath, "docker://") {
		// download docker image
	} else {
		if strings.HasPrefix(userInputPath, "http") {
			// download image from http
		} else {
			// U
			mydisk, err = stdpull.NewStdDiskPull(userInputPath, imagePath)
		}
	}
	if err != nil {
		return err
	}
	return mydisk.Get()
}
