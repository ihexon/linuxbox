package decompress

import (
	"bauklotze/pkg/machine/define"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	decompressedFileFlag = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
)

// decompressedFilePath is a full path with contain file name !!
func Decompress(compressedVMFile *define.VMFile, decompressedFilePath string) error {
	compressedFilePath := compressedVMFile.GetPath()
	var (
		d   decompressor
		err error
	)
	if d, err = newDecompressor(compressedFilePath); err != nil {
		return err
	}
	return runDecompression(d, decompressedFilePath)
}

func newDecompressor(compressedFilePath string) (decompressor, error) {
	return newGenericDecompressor(compressedFilePath)
}

func newGenericDecompressor(compressedFilePath string) (*genericDecompressor, error) {
	d := &genericDecompressor{}
	d.compressedFilePath = compressedFilePath
	stat, err := os.Stat(d.compressedFilePath)
	if err != nil {
		return nil, err
	}
	d.compressedFileInfo = stat
	return d, nil
}

func runDecompression(d decompressor, decompressedFilePath string) (retErr error) {
	compressedFileReader, err := d.compressedFileReader()
	if err != nil {
		return err
	}
	defer d.close()
	//filesize := d.compressedFileSize()
	var decompressedFileWriter *os.File

	if decompressedFileWriter, err = os.OpenFile(decompressedFilePath, decompressedFileFlag, d.compressedFileMode()); err != nil {
		logrus.Errorf("Unable to open destination file %s for writing: %q", decompressedFilePath, err)
		return err
	}

	defer func() {
		if err := decompressedFileWriter.Close(); err != nil {
			logrus.Warnf("Unable to to close destination file %s: %q", decompressedFilePath, err)
			if retErr == nil {
				retErr = err
			}
		}
	}()

	if err = d.decompress(decompressedFileWriter, compressedFileReader); err != nil {
		logrus.Errorf("Error extracting compressed file: %q", err)
		return err
	}
	return nil
}
