package stdpull

import (
	"bauklotze/pkg/archiver"
	fileutils "bauklotze/pkg/ioutils"
	"bauklotze/pkg/machine/define"
	"github.com/sirupsen/logrus"
)

type StdDiskPull struct {
	inputPath *define.VMFile
	finalPath *define.VMFile
}

func (s *StdDiskPull) Get() error {
	if err := fileutils.Exists(s.inputPath.GetPath()); err != nil {
		// could not find user input disk
		return err
	}
	// 解压 rootfs 或者镜像到 .local/share/containers/podman/machine/[vmType]/podman-machine-default-amd64 下
	logrus.Debugf("decompressing (if needed) %s to %s", s.inputPath.GetPath(), s.finalPath.GetPath())

	return archiver.Decompress(s.inputPath, s.finalPath.GetPath())
}

// userInputPath 是用户指定的 rootfs 路径
// finalpath 实际上是本地存储的 rootfs 的路径
// 将 userInputPath 解压到 finalpath
func NewStdDiskPull(inputPath string, finalpath *define.VMFile) (*StdDiskPull, error) {
	inputImage, err := define.NewMachineFile(inputPath)
	if err != nil {
		return nil, err
	}
	return &StdDiskPull{inputPath: inputImage, finalPath: finalpath}, nil
}
