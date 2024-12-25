//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package vfkit

import (
	"fmt"

	"bauklotze/pkg/machine/vmconfigs"

	gvproxy "github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/sirupsen/logrus"
)

// StartGenericNetworking most logic has been removed
func StartGenericNetworking(mc *vmconfigs.MachineConfig, cmd *gvproxy.GvproxyCommand) error {
	gvProxySock, err := mc.GVProxySocket()
	if err != nil {
		return fmt.Errorf("failed to get gvproxy socket: %w", err)
	}
	// make sure it does not exist before gvproxy is called
	logrus.Infof("Deleting gvproxy socket %s", gvProxySock.GetPath())
	if err := gvProxySock.Delete(); err != nil {
		return fmt.Errorf("failed to delete gvproxy socket: %w", err)
	}

	cmd.AddVfkitSocket(fmt.Sprintf("unixgram://%s", gvProxySock.GetPath()))

	return nil
}
