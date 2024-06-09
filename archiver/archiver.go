package archiver

import (
	"fmt"
	"io"
)

type Decompressor interface {
	Decompress(in io.Reader, out io.Writer) error
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
