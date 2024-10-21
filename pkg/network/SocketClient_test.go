package network

import (
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestHttpClient(t *testing.T) {
	connCtx, err := NewConnection("unix:///tmp/report_url.socks")
	connCtx.Headers = http.Header{
		"Content-Type": []string{"application/json"},
	}
	connCtx.UrlParameter = url.Values{
		"key": []string{"value"},
	}

	//connCtx.Body = strings.NewReader("Hello, World!")
	response, err := connCtx.DoRequest("GET", "notify")
	if err != nil {
		t.Errorf(err.Error())
	}

	if response.Response != nil {
		body, _ := io.ReadAll(response.Response.Body)
		t.Logf("Response Body: %s", string(body))
		defer response.Response.Body.Close()
	}
}
