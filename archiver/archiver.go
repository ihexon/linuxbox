package archiver

import (
	"fmt"
	"io"
	"os"
)

// 所有的 压缩实现 都需要实现Decompress方法
type Decompressor interface {
	Decompress(in io.Reader, out io.Writer) error
}

// 基于文件的 IO
func DecompressFile(source, destination string, overwrite bool) error {

	Iface, err := ByExt(source)
	if err != nil {
		return err
	}
	c, ok := Iface.(Decompressor)
	if !ok {
		return fmt.Errorf("format specified by source filename is not a recognized compression algorithm: %s", source)
	}
	return FileCompressor{Decompressor: c}.DecompressFile(source, destination, overwrite)
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
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

// ExtensionChecker 是一个（方法）数组，每个方法在 ByExt 都有机会被执行一次
var extCheckers = []ExtensionChecker{
	&Zstd{},
	&Xz{},
}

// 注意：
// FileCompressor 是基于文件的 IO: ovm --import rootfs.tar.xz
// StreamCompressor 是基于流的 IO: cat rootfs | ovm --import
// StreamCompressor 未实现
type FileCompressor struct {
	Decompressor
	// Whether to overwrite existing files when creating files.
	OverwriteExisting bool
}

func (fc FileCompressor) DecompressFile(source, destination string, overwrite bool) error {
	if fc.Decompressor == nil {
		return fmt.Errorf("no decompressor specified")
	}
	fc.OverwriteExisting = overwrite
	if fileExists(destination) && !fc.OverwriteExisting {
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
