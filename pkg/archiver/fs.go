package archiver

import "io"

type compressorCloser interface {
	io.Closer
	closeCompressor(io.Closer)
}
