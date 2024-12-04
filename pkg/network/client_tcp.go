//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package network

import (
	"context"
	"net"
	"net/http"
)

func tcpClient(myConn *Connection) (*http.Client, error) { //nolint:unparam
	dialContext := func(ctx context.Context, _, _ string) (net.Conn, error) {
		return net.Dial("tcp", myConn.URI.Host)
	}
	myConn.TCPClient = &http.Client{
		Transport: &http.Transport{
			DialContext:        dialContext,
			DisableCompression: true,
		},
	}

	return myConn.TCPClient, nil
}
