package archiver

import (
	"fmt"
	"github.com/klauspost/compress/zstd"
	"io"
	"path/filepath"
)

var zstdHeader = []byte{0x28, 0xb5, 0x2f, 0xfd}

type Zstd struct {
	DecoderOptions []zstd.DOption
}

func NewZstd() *Zstd {
	return new(Zstd)
}

func (zs *Zstd) Decompress(in io.Reader, out io.Writer) error {
	readed, _ := zs.openReader(in)
	defer readed.Close()
	_, err := io.Copy(out, readed)
	return err
}

// // 打开一个 zstd 输入流
func (zs Zstd) openReader(in io.Reader) (io.ReadCloser, error) {
	zr, err := zstd.NewReader(in, zs.DecoderOptions...)
	if err != nil {
		return nil, err
	}
	// 解压的本质是对流的处理，Close 方法什么都不做
	return io.NopCloser(zr), nil
}

func (zs *Zstd) CheckExt(filename string) error {
	if filepath.Ext(filename) != ".zst" {
		return fmt.Errorf("filename must have a .zst extension")
	}
	return nil
}
