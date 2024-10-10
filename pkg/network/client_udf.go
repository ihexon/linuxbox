package network

import (
	"context"
	"net"
	"net/http"
)

func unixClient(myConnection *Connection) *http.Client {
	myConnection.UnixClient = &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", myConnection.URI.Path)
			},
			DisableCompression: true,
		},
	}

	return myConnection.UnixClient
}
