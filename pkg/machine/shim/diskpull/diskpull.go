package diskpull

import (
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/shim/stdpull"
	"bauklotze/pkg/ovmdisk"
)

// userInputImageFilePath 是用户指定的 rootfs 路径
// mc.ImagePath 实际上是 rootfs 的路径
// 先解压用户输入的 userInputPath，解压到 imagePath
func GetDisk(userInputImageFilePath string, dirs *define.MachineDirs, imageFilePath *define.VMFile, vmType define.VMType, name string) error {

	var (
		err    error
		mydisk ovmdisk.Disker
	)
	// 填充 StdDiskPull 结构体
	mydisk, err = stdpull.NewStdDiskPull(userInputImageFilePath, imageFilePath)

	if err != nil {
		return err
	}
	return mydisk.Get()
}
