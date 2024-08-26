package apple

import (
	"bauklotze/pkg/machine/vmconfigs"
	strongunits "bauklotze/pkg/storage"
	"github.com/sirupsen/logrus"
	"os"
)

func ResizeDisk(mc *vmconfigs.MachineConfig, newSize strongunits.GiB) error {
	logrus.Debugf("resizing %s to %d bytes", mc.ImagePath.GetPath(), newSize.ToBytes())
	return os.Truncate(mc.ImagePath.GetPath(), int64(newSize.ToBytes()))
}
