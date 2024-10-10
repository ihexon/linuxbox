package network

import (
	"context"
	"net"
	"net/http"
)

func tcpClient(myConn *Connection) (*http.Client, error) {
	dialContext := func(ctx context.Context, _, _ string) (net.Conn, error) {
		return net.Dial("tcp", myConn.URI.Host)
	}
	myConn.TcpClient = &http.Client{
		Transport: &http.Transport{
			DialContext:        dialContext,
			DisableCompression: true,
		},
	}

	return myConn.TcpClient, nil
}
