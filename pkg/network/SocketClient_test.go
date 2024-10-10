package network

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestHttpClient(t *testing.T) {

	connCtx, err := NewConnection("tcp://127.0.0.1:8080")
	connCtx.headers = http.Header{
		"Content-Type": []string{"application/json"},
	}
	connCtx.urlParameter = url.Values{
		"key": []string{"value"},
	}
	dataReader := strings.NewReader("Hello, World!")
	response, err := connCtx.DoRequest("POST", "/1/2/4/5/name", dataReader)
	if err != nil {
		t.Errorf(err.Error())
	}

	if response.Response != nil {
		body, _ := io.ReadAll(response.Response.Body)
		t.Logf("Response Body: %s", string(body))
		defer response.Response.Body.Close()
	}
}
