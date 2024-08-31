package stdpull

import (
	"bauklotze/pkg/archiver/decompress"
	"bauklotze/pkg/machine/machineDefine"
	"github.com/containers/storage/pkg/fileutils"
	"github.com/sirupsen/logrus"
)

type StdDiskPull struct {
	// all machineDefine.VMFile are not dir instead the full path contained file name
	inputPath *machineDefine.VMFile
	finalPath *machineDefine.VMFile
}

func NewStdDiskPull(inputPath string, finalpath *machineDefine.VMFile) (*StdDiskPull, error) {
	inputImage, err := machineDefine.NewMachineFile(inputPath)
	if err != nil {
		return nil, err
	}
	return &StdDiskPull{inputPath: inputImage, finalPath: finalpath}, nil
}

// Get StdDiskPull: Get just decompress the `inputPath *machineDefine.VMFile` to `finalPath *machineDefine.VMFile`
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
