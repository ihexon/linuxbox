//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package network

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type APIResponse struct {
	*http.Response
	Request *http.Request
}

var myConnection = &Conn{}

type Conn struct {
	URI          *url.URL
	UnixClient   *http.Client
	URLParameter url.Values
	Headers      http.Header
	Body         io.Reader
}

const defaultTimeout = 100 * time.Millisecond

func (c *Conn) Request(httpMethod, endpoint string) (*APIResponse, error) {
	var (
		err      error
		response *http.Response
		client   *http.Client
		baseURL  string
	)

	if c.URI.Scheme == "unix" {
		// Allow path prefixes for tcp connections to match Docker behavior
		baseURL = "http://local"
		client = c.UnixClient
	}

	uri := fmt.Sprintf("%s/%s", baseURL, endpoint)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, httpMethod, uri, c.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if len(c.URLParameter) > 0 {
		req.URL.RawQuery = c.URLParameter.Encode()
	}

	for key, val := range c.Headers {
		for _, v := range val {
			req.Header.Add(key, v)
		}
	}

	response, err = client.Do(req) //nolint:bodyclose
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	return &APIResponse{response, req}, nil
}

// JoinURL elements with '/'
func JoinURL(elements ...string) string {
	return "/" + strings.Join(elements, "/")
}

func NewConn(uri string) (*Conn, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("not a valid url: %s: %w", uri, err)
	}
	myConnection.URI = u

	switch u.Scheme {
	case "unix":
		if !strings.HasPrefix(uri, "unix:///") {
			// autofix unix://path_element vs unix:///path_element
			u.Path = JoinURL(u.Host, u.Path)
			u.Host = ""
		}
		myConnection.URI = u
		myConnection.UnixClient = unixClient(myConnection)
	default:
		return nil, fmt.Errorf("unable to create connection. %q is not a supported schema", u.Scheme)
	}
	return myConnection, nil
}
