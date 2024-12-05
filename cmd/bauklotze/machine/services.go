//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package machine

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"runtime"

	"bauklotze/cmd/bauklotze/validata"
	"bauklotze/cmd/registry"
	"bauklotze/pkg/api/server"
	"bauklotze/pkg/machine/env"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
	listenURL, err := resolveAPIURI(args)
	if err != nil {
		return fmt.Errorf("%s is an invalid socket destination: %w", args[0], err)
	}

	if err := server.RestService(context.Background(), listenURL); err != nil {
		return fmt.Errorf("failed to start API service: %w", err)
	}

	return nil
}

// resolveAPIURI resolves the API URI from the given arguments, if no arguments are given, it tries to get the URI from the env.DefaultRootAPIAddress
func resolveAPIURI(uri []string) (*url.URL, error) {
	apiuri := env.DefaultRootAPIAddress

	if len(uri) > 0 && uri[0] != "" {
		apiuri = uri[0]
	}
	logrus.Infof("%s @ try listen URI: %s", runtime.FuncForPC(reflect.ValueOf(resolveAPIURI).Pointer()).Name(), apiuri)
	url, err := url.Parse(apiuri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI: %s, %w", apiuri, err)
	}

	return url, nil
}
