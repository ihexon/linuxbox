//go:build amd64 || arm64

package connection

import (
	"bauklotze/pkg/config"
	"bauklotze/pkg/machine/define"
	"errors"
	"github.com/sirupsen/logrus"
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

func addConnection(cons []connection, identity string, isDefault bool) error {
	if len(identity) < 1 {
		return errors.New("identity must be defined")
	}

	return config.EditConnectionConfig(func(cfg *config.ConnectionsFile) error {
		for i, con := range cons {
			//if _, ok := cfg.Connection.Connections[con.name]; ok {
			//	return fmt.Errorf("cannot overwrite connection %q", con.name)
			//}
			
			dst := config.Destination{
				URI:      con.uri.String(),
				Identity: identity,
			}

			if isDefault && i == 0 {
				cfg.Connection.Default = con.name
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
	})
}

func UpdateConnectionPairPort(name string, port, uid int, remoteUsername string, identityPath string) error {
	cons := createConnections(name, uid, port, remoteUsername)
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

func RemoveConnections(machines map[string]bool, names ...string) error {
	var dest config.Destination
	var service string

	if err := config.EditConnectionConfig(func(cfg *config.ConnectionsFile) error {
		err := setNewDefaultConnection(cfg, &dest, &service, names...)
		if err != nil {
			return err
		}

		rootful, ok := machines[service]
		if ok {
			updateConnection(cfg, rootful, service, service+"-root")
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func updateConnection(cfg *config.ConnectionsFile, rootful bool, name, rootfulName string) {
	if name == cfg.Connection.Default && rootful {
		cfg.Connection.Default = rootfulName
	} else if rootfulName == cfg.Connection.Default && !rootful {
		cfg.Connection.Default = name
	}
}

// setNewDefaultConnection iterates through the available system connections and
// sets the first available connection as the new default
func setNewDefaultConnection(cfg *config.ConnectionsFile, dest *config.Destination, service *string, names ...string) error {
	// delete the connection associated with the names and if that connection is
	// the default, reset the default connection
	for _, name := range names {
		if _, ok := cfg.Connection.Connections[name]; ok {
			delete(cfg.Connection.Connections, name)
		} else {
			logrus.Warnf("unable to find connection named %q", name)
		}

		if cfg.Connection.Default == name {
			cfg.Connection.Default = ""
		}
	}

	// If there is a podman-machine-default system connection, immediately set that as the new default
	if c, ok := cfg.Connection.Connections[define.DefaultMachineName]; ok {
		cfg.Connection.Default = define.DefaultMachineName
		*dest = c
		*service = define.DefaultMachineName
		return nil
	}

	// set the new default system connection to the first in the map
	for con, d := range cfg.Connection.Connections {
		cfg.Connection.Default = con
		*dest = d
		*service = con
		break
	}
	return nil
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
