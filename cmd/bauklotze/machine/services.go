package machine

import (
	"bauklotze/cmd/bauklotze/validata"
	"bauklotze/cmd/registry"
	"bauklotze/pkg/api/server"
	"bauklotze/pkg/machine/env"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"os"
)

var (
	srvDescription = `Run an API service

Enable a listening service for API access to Podman commands.
`
	serviceCmd = &cobra.Command{
		Use:               "service [options] [URI]",
		Args:              cobra.MaximumNArgs(1),
		Short:             "Run API service",
		Long:              srvDescription,
		RunE:              service,
		ValidArgsFunction: validata.AutocompleteDefaultOneArg,
		Example:           `bauklotze system service tcp://127.0.0.1:8888`,
	}
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: serviceCmd,
		Parent:  systemCmd,
	})
}

func service(cmd *cobra.Command, args []string) error {
	apiurl, _ := resolveAPIURI(args)
	if len(apiurl) > 0 {
		uri, err := url.Parse(apiurl)
		if err != nil {
			return err
		}

		// We do not support unix socket file as api endpoint now
		if uri.Scheme == "unix" {
			return fmt.Errorf("we do not support unix socket file as api endpoint now")
		}

	}
	return server.RestService(cmd.Flags(), apiurl)
}

// TODO: Support unix socket as rest API
// resolveAPIURI resolves the API URI from the given arguments, if no arguments are given, it tries to get the URI from the env.DefaultRootAPIAddress
func resolveAPIURI(uri []string) (string, error) {
	// If given no api addr, try to get restapi addr from environment
	if len(uri) == 0 {
		if v, found := os.LookupEnv(env.DefaultRootAPIEnv); found {
			logrus.Infof("System environment %s: %q used to determine API endpoint", env.DefaultRootAPIEnv, v)
			uri = []string{v}
		}
	}

	switch {
	case len(uri) > 0 && uri[0] != "":
		return uri[0], nil
	default:
		// Default return tcp://127.0.0.1:65176
		return env.DefaultRootAPIAddress, nil
	}
}
