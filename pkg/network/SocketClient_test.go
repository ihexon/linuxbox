package network

import (
	"context"
	"net/http"
	"net/url"
	"testing"
)

func TestHttpClient(t *testing.T) {

	ctx := context.Background()

	params := url.Values{}
	params.Add("param", "I'M GOING TO KILL MYSELF")

	headers := http.Header{}
	headers.Add("Header", "FUCK EVERY ONE & FUCK ME")

	connCtx, err := NewConnection(ctx, "tcp://127.0.0.1:8080")
	if err != nil {
		return
	}

	client, err := GetClient(connCtx)
	if err != nil {
		return
	}

	response, err := client.DoRequest(ctx, "POST", "/1/2/4/5/name", params, headers)
	if err != nil {
		t.Errorf(err.Error())
	}

	connCtx2, err := NewConnection(ctx, "tcp://127.0.0.1:8080")
	client, err = GetClient(connCtx2)
	if err != nil {
		return
	}

	response, err = client.DoRequest(ctx, "GET", "/15/name", params, headers)
	if err != nil {
		t.Errorf(err.Error())
	}

	connCtx3, err := NewConnection(ctx, "unix:///tmp/zzh.sock")
	client, err = GetClient(connCtx3)
	if err != nil {
		return
	}

	response, err = client.DoRequest(ctx, "GET", "/15/name", params, headers)
	if err != nil {
		t.Errorf(err.Error())
	}

	defer response.Body.Close()
}
