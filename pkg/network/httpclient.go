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
	"time"
)

type Connection struct {
	URI          *url.URL
	TcpClient    *http.Client
	UnixClient   *http.Client
	UrlParameter url.Values
	Headers      http.Header
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
	case "tcp":
		if !strings.HasPrefix(uri, "tcp://") {
			return myConnection, errors.New("tcp URIs should begin with tcp://")
		}
		myConnection.URI = _url
		myConnection.TcpClient, err = tcpClient(myConnection)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unable to create connection. %q is not a supported schema", _url.Scheme)
	}
	return myConnection, nil
}

func (c *Connection) DoRequest(httpMethod, endpoint string, httpBody io.Reader) (*APIResponse, error) {
	var (
		err      error
		response *http.Response
		client   *http.Client
	)

	baseURL := ""
	if c.URI.Scheme == "tcp" || c.URI.Scheme == "http" {
		// Allow path prefixes for tcp connections to match Docker behavior
		baseURL = "http://" + c.URI.Host + c.URI.Path
		client = c.TcpClient
	}

	if c.URI.Scheme == "unix" {
		// Allow path prefixes for tcp connections to match Docker behavior
		baseURL = "http://local/"
		client = c.UnixClient
	}

	uri := fmt.Sprintf(baseURL + "/" + endpoint)
	logrus.Infof("DoRequest Method: %s URI: %v", httpMethod, uri)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, httpMethod, uri, httpBody)
	if err != nil {
		return nil, err
	}
	if len(c.UrlParameter) > 0 {
		req.URL.RawQuery = c.UrlParameter.Encode()
	}

	for key, val := range c.Headers {
		for _, v := range val {
			req.Header.Add(key, v)
		}
	}

	response, err = client.Do(req)
	return &APIResponse{response, req}, err
}
