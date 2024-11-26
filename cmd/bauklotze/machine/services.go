package machine

import (
	"bauklotze/cmd/bauklotze/validata"
	"bauklotze/cmd/registry"
	"bauklotze/pkg/api/server"
	"bauklotze/pkg/machine/env"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"reflect"
	"runtime"
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
		PersistentPreRunE: machinePreRunE,
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
	listenUrl, err := resolveAPIURI(args)
	if err != nil {
		return fmt.Errorf("%s is an invalid socket destination", args[0])
	}
	return server.RestService(context.Background(), listenUrl)
}

// resolveAPIURI resolves the API URI from the given arguments, if no arguments are given, it tries to get the URI from the env.DefaultRootAPIAddress
func resolveAPIURI(uri []string) (*url.URL, error) {
	apiuri := env.DefaultRootAPIAddress

	if len(uri) > 0 && uri[0] != "" {
		apiuri = uri[0]
	}
	logrus.Infof("%s @ try listen URI: %s", runtime.FuncForPC(reflect.ValueOf(resolveAPIURI).Pointer()).Name(), apiuri)
	return url.Parse(apiuri)
}
