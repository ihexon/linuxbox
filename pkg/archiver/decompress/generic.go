package decompress

import (
	"github.com/mholt/archiver/v4"
	"github.com/sirupsen/logrus"
	"io"
	"io/fs"
	"os"
)

type genericDecompressor struct {
	compressedFilePath string
	compressedFile     *os.File
	compressedFileInfo os.FileInfo
}

func (d *genericDecompressor) compressedFileMode() fs.FileMode {
	return d.compressedFileInfo.Mode()
}

func (d *genericDecompressor) compressedFileSize() int64 {
	return d.compressedFileInfo.Size()
}

func (d *genericDecompressor) compressedFileReader() (io.ReadCloser, error) {
	compressedFile, err := os.Open(d.compressedFilePath)
	if err != nil {
		return nil, err
	}
	d.compressedFile = compressedFile
	return compressedFile, nil
}

func (d *genericDecompressor) close() {
	if err := d.compressedFile.Close(); err != nil {
		logrus.Errorf("Unable to close compressed file: %q", err)
	}
}

func (d *genericDecompressor) decompress(w io.WriteSeeker, r io.Reader) error {
	// If stream is non-nil then the returned io.Reader will always be non-nil and will read from the same point as the reader which was passed in;
	// It should be used in place of the input stream after calling Identify() because it preserves and re-reads the bytes that were already read during the identification process
	format, reader, err := archiver.Identify("", r)
	if err != nil {
		return err
	}

	decompressorProvider, ok := format.(archiver.Decompressor)

	if ok {
		c, err := decompressorProvider.OpenReader(reader)
		defer func() {
			if err := c.Close(); err != nil {
				logrus.Errorf("generic decompressor: unable to close file: %q", err)
			}
		}()

		if err != nil {
			return err
		}
		_, err = io.Copy(w, c)
	}
	return err
}
