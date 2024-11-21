//go:build amd64 || arm64

package machine

import (
	whatProvider "bauklotze/pkg/machine/provider"
	"bauklotze/pkg/machine/vmconfigs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var provider vmconfigs.VMProvider

func machinePreRunE(cmd *cobra.Command, args []string) error {
	var err error
	logrus.Infof("Try to get current hypervisor provider...")
	provider, err = whatProvider.Get()
	if err != nil {
		return err
	}
	logrus.Infof("Got current hypervisor provider %s", provider.VMType().String())
	return err
}

//func closeMachineEvents(cmd *cobra.Command, _ []string) error {
//	return nil
//}
