//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package vmconfigs

import (
	"fmt"

	"bauklotze/pkg/machine/define"
)

func readySocket(name string, machineRuntimeDir *define.VMFile) (*define.VMFile, error) {
	socketName := fmt.Sprintf("%s-ready.sock", name)
	return machineRuntimeDir.AppendToNewVMFile(socketName, nil)
}

func gvProxySocket(name string, machineRuntimeDir *define.VMFile) (*define.VMFile, error) {
	socketName := fmt.Sprintf("%s-gvproxy.sock", name)
	return machineRuntimeDir.AppendToNewVMFile(socketName, nil)
}

func podmanAPISocketOnHost(name string, socketDir *define.VMFile) (*define.VMFile, error) {
	socketName := fmt.Sprintf("%s-podman-api.sock", name)
	return socketDir.AppendToNewVMFile(socketName, nil)
}

func ignitionSocket(name string, socketDir *define.VMFile) (*define.VMFile, error) {
	socketName := fmt.Sprintf("%s-ignition.sock", name)
	return socketDir.AppendToNewVMFile(socketName, nil)
}
