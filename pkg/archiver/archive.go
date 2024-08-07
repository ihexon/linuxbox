package archiver

import (
	"bauklotze/pkg/machine/define"
	"context"
	"github.com/mholt/archiver/v4"
	"github.com/sirupsen/logrus"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	fs.FileInfo
	Header        interface{}
	NameInArchive string
	LinkTarget    string
	Open          func() (io.ReadCloser, error)
}

type FileHandler func(ctx context.Context, f File) error

// Decompress : Departed, will be delete next version
func Decompress(compressedVMFile *define.VMFile, targetPathStr string) error {
	compressedFileNameWithFullPath := compressedVMFile.GetPath()
	file, err := os.OpenFile(compressedFileNameWithFullPath, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	format, reader, err := Identify(compressedVMFile.GetPath(), file)
	if err != nil {
		return err
	}
	decom, ok := format.(archiver.Decompressor)
	if ok {
		ioReadCloser, err := decom.OpenReader(reader)
		if err != nil {
			return err
		}
		defer ioReadCloser.Close()
		outputDir := filepath.Join(targetPathStr)
		err = os.MkdirAll(outputDir, 0755)
		if err != nil {
			return err
		}
		decompressedFilename := strings.TrimSuffix(filepath.Base(compressedFileNameWithFullPath), format.Name())
		outputFile, err := os.Create(filepath.Join(outputDir, decompressedFilename))
		if err != nil {
			return err
		}
		defer outputFile.Close()
		if wed, err := io.Copy(outputFile, ioReadCloser); err != nil {
			return err
		} else {
			logrus.Debugf("Decompress bytes: %d from %s to %s ", wed, compressedFileNameWithFullPath, decompressedFilename)
		}
	}
	return err
}
