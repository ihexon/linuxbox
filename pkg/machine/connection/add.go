//go:build amd64 || arm64

package connection

import (
	"bauklotze/pkg/machine/define"
	"strconv"
)

// AddSSHConnectionsToPodmanSocket adds SSH connections to the podman socket if
// no ignition path is provided
func AddSSHConnectionsToPodmanSocket(uid, port int, identityPath, name, remoteUsername string, opts define.InitOptions) error {
	cons := createConnections(name, uid, port, remoteUsername)
	return addConnection(cons, identityPath, true)
}

func createConnections(name string, uid, port int, remoteUsername string) []connection {
	uriRoot := makeSSHURL(LocalhostIP, "/run/podman/podman.sock", strconv.Itoa(port), "root")

	return []connection{
		{
			name: name + "-root",
			uri:  uriRoot,
		},
	}
}
