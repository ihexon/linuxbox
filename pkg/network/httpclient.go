//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package network

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Connection struct {
	URI          *url.URL
	UnixClient   *http.Client
	URLParameter url.Values
	Headers      http.Header
	Body         io.Reader
}

var myConnection = &Connection{}

type APIResponse struct {
	*http.Response
	Request *http.Request
}

// JoinURL elements with '/'
func JoinURL(elements ...string) string {
	return "/" + strings.Join(elements, "/")
}

func NewConnection(uri string) (*Connection, error) {
	_url, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("not a valid url: %s: %w", uri, err)
	}
	myConnection.URI = _url

	switch _url.Scheme {
	case "unix":
		if !strings.HasPrefix(uri, "unix:///") {
			// autofix unix://path_element vs unix:///path_element
			_url.Path = JoinURL(_url.Host, _url.Path)
			_url.Host = ""
		}
		myConnection.URI = _url
		myConnection.UnixClient = unixClient(myConnection)
	default:
		return nil, fmt.Errorf("unable to create connection. %q is not a supported schema", _url.Scheme)
	}
	return myConnection, nil
}

const defaultTimeout = 100 * time.Millisecond

func (c *Connection) DoRequest(httpMethod, endpoint string) (*APIResponse, error) {
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

func (o *OvmJSListener) SendEventToOvmJs(event, message string) {
	if o.ReportURL == "" {
		return
	}
	connCtx, err := NewConnection(o.ReportURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "report url not valid, only support unix:/// or tcp:// proto\n")
		return
	}

	connCtx.Headers = http.Header{
		"Content-Type": []string{PlainTextContentType},
	}
	connCtx.URLParameter = url.Values{
		"event":   []string{event},
		"message": []string{message},
	}
	logrus.Infof("Send Event to %s , %s", connCtx.URI, connCtx.URLParameter)
	req, err := connCtx.DoRequest("GET", "notify")
	if err != nil {
		logrus.Warnf("Failed to notify %q: %v\n", o.ReportURL, err)
	} else {
		req.Body.Close()
	}
}

var (
	Reporter OvmJSListener
	once     sync.Once
)

func NewReporter(url string) *OvmJSListener {
	once.Do(func() {
		Reporter = OvmJSListener{
			ReportURL: url,
		}
	})
	return &Reporter
}
