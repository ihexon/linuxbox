package machine

import "github.com/spf13/cobra"

var (
	srvDescription = `Run an API service

Enable a listening service for API access to Podman commands.
`

	srvCmd = &cobra.Command{
		Use:     "apiserv [options] [URI]",
		Args:    cobra.MaximumNArgs(1),
		Short:   "Run API service",
		Long:    srvDescription,
		RunE:    service,
		Example: `podman system apiserv  tcp://localhost:8888`,
	}
)

// TODO
func service(cmd *cobra.Command, args []string) error {
	return nil

}
