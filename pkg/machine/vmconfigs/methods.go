package vmconfigs

import (
	"bauklotze/pkg/machine/define"
	"encoding/json"
	"fmt"
	"github.com/containers/storage/pkg/ioutils"
)

// write is a non-locking way to write the machine configuration file to disk
func (mc *MachineConfig) Write() error {
	if mc.ConfigPath == nil {
		return fmt.Errorf("no configuration file associated with vm %q", mc.Name)
	}
	b, err := json.Marshal(mc)
	if err != nil {
		return err
	}
	return ioutils.AtomicWriteFile(mc.ConfigPath.GetPath(), b, define.DefaultFilePerm)
}
