package archiver

import (
	"context"
	"io"
	"io/fs"
)

type File struct {
	fs.FileInfo

	// The file header as used/provided by the archive format.
	// Typically, you do not need to set this field when creating
	// an archive.
	Header interface{}

	// The path of the file as it appears in the archive.
	// This is equivalent to Header.Name (for most Header
	// types). We require it to be specified here because
	// it is such a common field and we want to preserve
	// format-agnosticism (no type assertions) for basic
	// operations.
	//
	// EXPERIMENTAL: If inserting a file into an archive,
	// and this is left blank, the implementation of the
	// archive format can default to using the file's base
	// name.
	NameInArchive string

	// For symbolic and hard links, the target of the link.
	// Not supported by all archive formats.
	LinkTarget string

	// A callback function that opens the file to read its
	// contents. The file must be closed when reading is
	// complete. Nil for files that don't have content
	// (such as directories and links).
	Open func() (io.ReadCloser, error)
}

type FileHandler func(ctx context.Context, f File) error
