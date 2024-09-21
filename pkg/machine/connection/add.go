//go:build amd64 || arm64

package connection

import (
	"bauklotze/pkg/machine/machineDefine"
	"strconv"
)

// AddSSHConnectionsToPodmanSocket adds SSH connections(host) to the podman socket(guest) if
// no ignition path is provided
func AddSSHConnectionsToPodmanSocket(uid, port int, identityPath, name, remoteUsername string, opts machineDefine.InitOptions) error {
	cons := createConnections(name, uid, port, remoteUsername)
	return addConnection(cons, identityPath, opts.IsDefault)
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
