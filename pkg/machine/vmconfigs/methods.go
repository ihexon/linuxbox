package vmconfigs

import (
	"bauklotze/pkg/machine/machineDefine"
	"encoding/json"
	"fmt"
	"github.com/containers/storage/pkg/ioutils"
	"github.com/sirupsen/logrus"
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
	logrus.Debugf("writing configuration file %q", mc.ConfigPath.Path)
	return ioutils.AtomicWriteFile(mc.ConfigPath.GetPath(), b, machineDefine.DefaultFilePerm)
}
