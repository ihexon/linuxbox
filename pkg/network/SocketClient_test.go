//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package network

import (
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestHttpClient(t *testing.T) {
	connCtx, err := NewConnection("unix:///tmp/report_url.socks")
	if err != nil {
		t.Error(err.Error())
	}
	connCtx.Headers = http.Header{
		"Content-Type": []string{"application/json"},
	}
	connCtx.URLParameter = url.Values{
		"key": []string{"value"},
	}

	// connCtx.Body = strings.NewReader("Hello, World!")
	response, err := connCtx.DoRequest("GET", "notify")
	if err != nil {
		t.Error(err.Error())
	}

	if response.Response != nil {
		body, _ := io.ReadAll(response.Response.Body)
		t.Logf("Response Body: %s", string(body))
		defer response.Response.Body.Close()
	}
}
