package decompress

import (
	"bauklotze/pkg/machine/define"
	"github.com/DataDog/zstd"
	"os"
)

func DecompressZstd(compressedFilePath *define.VMFile, decompressedFilePath *define.VMFile) error {
	var err error
	file, err := os.ReadFile(compressedFilePath.GetPath())
	if err != nil {
		return err
	}

	decompressData, err := zstd.Decompress(nil, file)
	if err != nil {
		return err
	}

	err = os.WriteFile(decompressedFilePath.GetPath(), decompressData, 0644)
	if err != nil {
		return err
	}

	return err
}
