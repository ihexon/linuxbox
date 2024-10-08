package network

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const clientKey = "myclient"

type Connection struct {
	URI    *url.URL
	Client *http.Client
}

type APIResponse struct {
	*http.Response
	Request *http.Request
}

// JoinURL elements with '/'
func JoinURL(elements ...string) string {
	return "/" + strings.Join(elements, "/")
}

var tcpConn Connection
var unixConn Connection

func NewConnection(ctx context.Context, uri string) (context.Context, error) {
	_url, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("not a valid url: %s: %w", uri, err)
	}

	switch _url.Scheme {
	case "unix":
		if !strings.HasPrefix(uri, "unix:///") {
			// autofix unix://path_element vs unix:///path_element
			_url.Path = JoinURL(_url.Host, _url.Path)
			_url.Host = ""
		}
		unixConn.URI = _url
		unixConn.Client = unixClient(unixConn)
		ctx = context.WithValue(ctx, clientKey, &unixConn)
	case "tcp":
		if !strings.HasPrefix(uri, "tcp://") {
			return nil, errors.New("tcp URIs should begin with tcp://")
		}
		tcpConn.URI = _url
		tcpConn.Client, err = tcpClient(tcpConn)
		if err != nil {
			return nil, err
		}
		ctx = context.WithValue(ctx, clientKey, &tcpConn)
	default:
		return nil, fmt.Errorf("unable to create connection. %q is not a supported schema", _url.Scheme)
	}

	return ctx, nil
}

func (c *Connection) DoRequest(ctx context.Context, httpMethod, endpoint string, queryParams url.Values, headers http.Header) (*APIResponse, error) {
	var (
		err      error
		response *http.Response
		httpBody io.Reader
	)

	baseURL := "http://d"
	if c.URI.Scheme == "tcp" {
		// Allow path prefixes for tcp connections to match Docker behavior
		baseURL = "http://" + c.URI.Host + c.URI.Path
	}
	uri := fmt.Sprintf(baseURL + "" + endpoint)
	logrus.Infof("DoRequest Method: %s URI: %v", httpMethod, uri)

	req, err := http.NewRequestWithContext(ctx, httpMethod, uri, httpBody)
	if err != nil {
		return nil, err
	}
	if len(queryParams) > 0 {
		req.URL.RawQuery = queryParams.Encode()
	}

	for key, val := range headers {
		for _, v := range val {
			req.Header.Add(key, v)
		}
	}

	// Give the Do three chances in the case of a comm/service hiccup
	response, err = c.Client.Do(req) //nolint:bodyclose // The caller has to close the body.

	return &APIResponse{response, req}, err
}

// GetClient from context build by NewConnection()
func GetClient(ctx context.Context) (*Connection, error) {
	if c, ok := ctx.Value(clientKey).(*Connection); ok {
		return c, nil
	}
	return nil, fmt.Errorf("%s not set in context", clientKey)
}
