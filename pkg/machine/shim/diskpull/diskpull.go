package diskpull

import (
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/shim/stdpull"
)

func GetDisk(userInputPath string, dirs *define.MachineDirs, imagePath *define.VMFile, vmType define.VMType, name string) error {
	// userInputPath 是用户指定的 rootfs 路径
	// mc.ImagePath 实际上是 rootfs 的路径
	mydisk, err := stdpull.NewStdDiskPull(userInputPath, imagePath)

	if err != nil {
		return err
	}
	return mydisk.Get()
}
