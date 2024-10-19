//go:build amd64 || arm64

package connection

import (
	"bauklotze/pkg/config"
	"errors"
	"fmt"
	"net"
	"net/url"
)

const (
	LocalhostIP    = "127.0.0.1"
	guestPodmanAPI = "/run/podman/podman.sock"
)

type connection struct {
	name string
	uri  *url.URL
}

func addConnection(cons []connection, identity string) error {
	if len(identity) < 1 {
		return errors.New("identity must be defined")
	}

	return config.EditConnectionConfig(
		func(cfg *config.ConnectionsFile) error {
			for _, con := range cons {
				dst := config.Destination{
					URI:      con.uri.String(),
					Identity: identity,
				}

				if cfg.Connection.Connections == nil {
					cfg.Connection.Connections = map[string]config.Destination{
						con.name: dst,
					}
					cfg.Connection.Default = con.name
				} else {
					cfg.Connection.Connections[con.name] = dst
				}
			}
			return nil
		},
	)
}

func UpdateConnectionPairPort(name string, port int, remoteUsername string, identityPath string) error {
	cons := createConnections(name, port, remoteUsername)
	return config.EditConnectionConfig(func(cfg *config.ConnectionsFile) error {
		for _, con := range cons {
			dst := config.Destination{
				URI:      con.uri.String(),
				Identity: identityPath,
			}
			cfg.Connection.Connections[con.name] = dst
		}

		return nil
	})
}

func RemoveConnections(names ...string) error {
	return config.EditConnectionConfig(func(cfg *config.ConnectionsFile) error {
		for _, name := range names {
			if _, ok := cfg.Connection.Connections[name]; ok {
				delete(cfg.Connection.Connections, name)
			} else {
				return fmt.Errorf("unable to find connection named %q", name)
			}

			if cfg.Connection.Default == name {
				cfg.Connection.Default = ""
			}
		}
		for service := range cfg.Connection.Connections {
			cfg.Connection.Default = service
			break
		}
		return nil
	})
}

// makeSSHURL creates a URL from the given input
func makeSSHURL(host, path, port, userName string) *url.URL {
	var hostname string
	if len(port) > 0 {
		hostname = net.JoinHostPort(host, port)
	} else {
		hostname = host
	}
	userInfo := url.User(userName)
	return &url.URL{
		Scheme: "ssh",
		User:   userInfo,
		Host:   hostname,
		Path:   path,
	}
}
