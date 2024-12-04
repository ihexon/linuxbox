//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build amd64 || arm64

package connection

import (
	"strconv"

	"bauklotze/pkg/machine/define"
)

// AddSSHConnectionsToPodmanSocket adds SSH connections to the podman socket if
// no ignition path is provided
func AddSSHConnectionsToPodmanSocket(uid, port int, identityPath, name, remoteUsername string, opts define.InitOptions) error {
	cons := createConnections(name, uid, port, remoteUsername)
	return addConnection(cons, identityPath, true)
}

func createConnections(name string, _uid, port int, _remoteUsername string) []connection {
	uriRoot := makeSSHURL(LocalhostIP, guestPodmanAPI, strconv.Itoa(port), "root")

	return []connection{
		{
			name: name + "-root",
			uri:  uriRoot,
		},
	}
}
