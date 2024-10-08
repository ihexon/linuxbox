//go:build amd64 || arm64

package connection

import (
	"bauklotze/pkg/machine/define"
	"strconv"
)

// AddSSHConnectionsToPodmanSocket
func AddSSHConnectionsToPodmanSocket(port int, identityPath, name, remoteUsername string, opts define.InitOptions) error {
	cons := createConnections(name, port, remoteUsername) // return a uriSSH connection
	return addConnection(cons, identityPath)
}

func createConnections(name string, port int, remoteUsername string) []connection {
	uriSSH := makeSSHURL(LocalhostIP, guestPodmanAPI, strconv.Itoa(port), remoteUsername)
	return []connection{
		{
			name: name + "-root",
			uri:  uriSSH,
		},
	}
}
