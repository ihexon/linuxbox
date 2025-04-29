//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package httpclient

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"bauklotze/pkg/machine/define"

	"github.com/sirupsen/logrus"
)

const defaultTimeout = 100 * time.Millisecond

type Client struct {
	baseURL string
	client  *http.Client
	ctx     context.Context

	Headers    http.Header
	Body       io.Reader
	QueryParam url.Values
}

func New() *Client {
	return &Client{
		baseURL:    "http://" + define.LocalHostURL,
		ctx:        context.Background(),
		Headers:    make(http.Header),
		QueryParam: make(url.Values),
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func (c *Client) SetBaseURL(url string) *Client {
	c.baseURL = strings.TrimRight(url, "/")
	return c
}

func (c *Client) SetTransport(transport http.RoundTripper) *Client {
	c.client.Transport = transport
	return c
}

func (c *Client) SetHeader(header, value string) *Client {
	c.Headers.Set(header, value)
	return c
}

func (c *Client) SetHeaders(headers map[string]string) *Client {
	for k, v := range headers {
		c.Headers.Set(k, v)
	}
	return c
}

func (c *Client) SetQueryParam(key, value string) *Client {
	c.QueryParam.Set(key, value)
	return c
}

func (c *Client) SetQueryParams(params map[string]string) *Client {
	for k, v := range params {
		c.SetQueryParam(k, v)
	}
	return c
}

func (c *Client) SetContext(ctx context.Context) {
	c.ctx = ctx
}

func (c *Client) Get(path string) error {
	uri := fmt.Sprintf("%s/%s", c.baseURL, strings.TrimLeft(path, "/"))
	req, err := http.NewRequestWithContext(c.ctx, http.MethodGet, uri, c.Body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.URL.RawQuery = c.QueryParam.Encode()
	req.Header = c.Headers

	response, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	logrus.Infof("Response Body: %s", string(body))
	defer response.Body.Close() //nolint:errcheck
	return nil
}

func CreateUnixTransport(path string) *http.Transport {
	u, _ := url.Parse(path)
	return &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", u.Path)
		},
		DisableCompression: true,
	}
}
