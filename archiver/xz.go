package archiver

import (
	"fmt"
	"github.com/ulikunitz/xz"
	"io"
	"path/filepath"
)

type Xz struct {
}

func NewXz() *Xz {
	return new(Xz)
}

func (me *Xz) Decompress(in io.Reader, out io.Writer) error {
	readed, _ := me.openReader(in)
	defer readed.Close()
	_, err := io.Copy(out, readed)
	return err
}

// 打开一个 XZ 输入流
func (me *Xz) openReader(r io.Reader) (io.ReadCloser, error) {
	r, err := xz.NewReader(r)
	if err != nil {
		return nil, err
	}
	// 解压的本质是对流的处理，Close 方法什么都不做
	return io.NopCloser(r), err
}

func (me *Xz) CheckExt(filename string) error {
	if filepath.Ext(filename) != ".xz" {
		return fmt.Errorf("filename must have a .xz extension")
	}
	return nil
}
