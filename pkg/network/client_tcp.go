package network

import (
	"context"
	"net"
	"net/http"
)

type TCPClient struct {
	address  string
	method   string
	response string
}

func tcpClient(_conn Connection) (*http.Client, error) {
	if _conn.Client == nil {
		dialContext := func(ctx context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("tcp", _conn.URI.Host)
		}
		_conn.Client = &http.Client{
			Transport: &http.Transport{
				DialContext:        dialContext,
				DisableCompression: true,
			},
		}
	}

	return _conn.Client, nil
}
