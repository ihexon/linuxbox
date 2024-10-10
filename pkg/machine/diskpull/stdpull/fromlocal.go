package stdpull

import (
	"bauklotze/pkg/archiver/decompress"
	"bauklotze/pkg/fileutils"
	"bauklotze/pkg/machine/define"
	"fmt"
	"github.com/sirupsen/logrus"
)

type StdDiskPull struct {
	// all define.VMFile are not dir instead the full path contained file name
	inputPath *define.VMFile
	finalPath *define.VMFile
}

func NewStdDiskPull(inputPath string, finalpath *define.VMFile) (*StdDiskPull, error) {
	inputImage, err := define.NewMachineFile(inputPath, nil)
	if err != nil {
		return nil, err
	}
	return &StdDiskPull{inputPath: inputImage, finalPath: finalpath}, nil
}

// Get StdDiskPull: Get just decompress the `inputPath *define.VMFile` to `finalPath *define.VMFile`
// Nothing interesting at all
func (s *StdDiskPull) Get() error {
	if err := fileutils.Exists(s.inputPath.GetPath()); err != nil {
		return fmt.Errorf("could not find user input disk: %w", err)
	}
	// 解压 rootfs 或者镜像解压成文件名为 .local/share/containers/podman/machine/[vmType]/podman-machine-default-amd64
	logrus.Infof("try to decompress %s to %s", s.inputPath.GetPath(), s.finalPath.GetPath())
	err := decompress.Decompress(s.inputPath, s.finalPath.GetPath())
	if err != nil {
		errors := fmt.Errorf("could not decompress %s to %s: %w", s.inputPath, s.finalPath.GetPath(), err)
		return errors
	}
	return nil
}
