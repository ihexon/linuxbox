package vmconfig

import (
	"bauklotze/pkg/machine/io"
	"errors"
	"time"
)

// ConfigDir is a simple helper to obtain the machine config dir
func (mc *MachineConfig) ConfigDir() (*io.FileWrapper, error) {
	if mc.Dirs == nil || mc.Dirs.ConfigDir == nil {
		return nil, errors.New("no configuration directory set")
	}
	return mc.Dirs.ConfigDir, nil
}

func (mc *MachineConfig) UpdateLastBoot() error { //nolint:unused
	mc.LastUp = time.Now()
	return mc.Write()
}
