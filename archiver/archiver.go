package archiver

import (
	"fmt"
	"io"
	"os"
)

type Decompressor interface {
	Decompress(in io.Reader, out io.Writer) error
}

func DecompressFile(source, destination string) error {
	cIface, err := ByExt(source)
	if err != nil {
		return err
	}
	c, ok := cIface.(Decompressor)
	if !ok {
		return fmt.Errorf("format specified by source filename is not a recognized compression algorithm: %s", source)
	}
	return FileCompressor{Decompressor: c}.DecompressFile(source, destination)
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func getFormat(filename string) (interface{}, error) {
	f, err := ByExt(filename)
	if err != nil {
		return nil, err
	}
	return f, nil
}

type Archiver struct {
	// Compress 未实现
	Decompressor
	OverwriteExisting bool
}

// 根据后缀，返回后缀名对应的类，如 xz 文件返回
// NewXz()，zstd 返回 NewZstd()
func ByExt(filename string) (interface{}, error) {
	var ifce interface{}
	for _, obj := range extCheckers {
		if err := obj.CheckExt(filename); err == nil {
			ifce = obj
			break
		}
	}
	switch ifce.(type) {
	case *Xz:
		return NewXz(), nil
	case *Zstd:
		return NewZstd(), nil
	}
	return nil, fmt.Errorf("format unrecognized by filename: %s", filename)
}

type ExtensionChecker interface {
	CheckExt(name string) error
}

// ExtensionChecker 是一个接口（方法）数组，每一个方法在 ByExt 都有机会被执行一次
var extCheckers = []ExtensionChecker{
	&Zstd{},
	&Xz{},
}

type FileCompressor struct {
	Decompressor
	// Whether to overwrite existing files when creating files.
	OverwriteExisting bool
}

// CompressFile reads the source file and compresses it to destination.
// The destination must have a matching extension.

// DecompressFile reads the source file and decompresses it to destination.
func (fc FileCompressor) DecompressFile(source, destination string) error {
	if fc.Decompressor == nil {
		return fmt.Errorf("no decompressor specified")
	}
	if !fc.OverwriteExisting && fileExists(destination) {
		return fmt.Errorf("file exists: %s", destination)
	}

	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	return fc.Decompress(in, out)
}
