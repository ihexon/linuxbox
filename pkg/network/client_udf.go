package network

import (
	"context"
	"net"
	"net/http"
)

func unixClient(_conn Connection) *http.Client {
	if _conn.Client == nil {
		_conn.Client = &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return (&net.Dialer{}).DialContext(ctx, "unix", _conn.URI.Path)
				},
				DisableCompression: true,
			},
		}
	}
	return _conn.Client
}
