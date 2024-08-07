package decompress

import (
	"io"
	"os"
)

type decompressor interface {
	compressedFileSize() int64
	compressedFileReader() (io.ReadCloser, error)
	decompress(w io.WriteSeeker, r io.Reader) error
	compressedFileMode() os.FileMode
	close()
}
