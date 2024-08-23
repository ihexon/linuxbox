package stdpull

import (
	"bauklotze/pkg/archiver/decompress"
	"bauklotze/pkg/machine/define"
	"github.com/containers/storage/pkg/fileutils"
	"github.com/sirupsen/logrus"
)

type StdDiskPull struct {
	// all define.VMFile are not dir, the full path contained file name
	inputPath *define.VMFile
	finalPath *define.VMFile
}

// 填充 StdDiskPull 结构体
func NewStdDiskPull(inputPath string, finalpath *define.VMFile) (*StdDiskPull, error) {
	inputImage, err := define.NewMachineFile(inputPath)
	if err != nil {
		return nil, err
	}
	return &StdDiskPull{inputPath: inputImage, finalPath: inputImage}, nil
}

// Get StdDiskPull: Get just decompress the `inputPath *define.VMFile` to `finalPath *define.VMFile`
// Nothing interesting at all
func (s *StdDiskPull) Get() error {
	if err := fileutils.Exists(s.inputPath.GetPath()); err != nil {
		// could not find user input disk
		return err
	}
	// 解压 rootfs 或者镜像解压成文件名为 .local/share/containers/podman/machine/[vmType]/podman-machine-default-amd64
	logrus.Debugf("decompressing (if needed) %s to %s", s.inputPath.GetPath(), s.finalPath.GetPath())
	return decompress.Decompress(s.inputPath, s.finalPath.GetPath())
}
