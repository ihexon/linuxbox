package decompress_v2

import (
	"bauklotze/pkg/machine/define"
	"fmt"
	"github.com/mholt/archiver/v4"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

func DecompressV2(compressedFilePath *define.VMFile, decompressedFilePath *define.VMFile) error {
	var err error
	inputFile, err := os.Open(compressedFilePath.GetPath())

	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	format, input, err := archiver.Identify("", inputFile)
	if err != nil {
		return fmt.Errorf("failed to identify format: %w", err)
	}
	decom, ok := format.(archiver.Decompressor)
	if !ok {
		return fmt.Errorf("unsupport format: %s", inputFile.Name())
	}

	c, err := decom.OpenReader(input)
	defer func() {
		if err = c.Close(); err != nil {
			logrus.Warnf("generic decompressor: unable to close file: %q", err)
		}
	}()
	if err != nil {
		return err
	}

	outputFile, err := os.OpenFile(decompressedFilePath.GetPath(), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	bytes, err := io.Copy(outputFile, c)
	logrus.Infof("Decompress %s %d bytes into %s", compressedFilePath.GetPath(), bytes, decompressedFilePath.GetPath())

	return err

}
