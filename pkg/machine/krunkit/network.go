//go:build darwin && arm64

package krunkit

import (
	"bauklotze/pkg/machine/vmconfigs"
	"fmt"
	gvproxy "github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/sirupsen/logrus"
)

// StartGenericNetworking is wrapped by apple provider methods
func StartGenericNetworking(mc *vmconfigs.MachineConfig, cmd *gvproxy.GvproxyCommand) error {
	gvProxySock, err := mc.GVProxySocket()
	if err != nil {
		return err
	}
	// make sure it does not exist before gvproxy is called
	if err := gvProxySock.Delete(); err != nil {
		logrus.Error(err)
	}
	cmd.AddVfkitSocket(fmt.Sprintf("unixgram://%s", gvProxySock.GetPath()))
	return nil
}
