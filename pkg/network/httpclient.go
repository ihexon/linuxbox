package network

import (
	"bauklotze/pkg/events"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ClientFactory struct {
	client *http.Client
}

func NewClientFactory(timeout time.Duration, endpoint string) *ClientFactory {
	var client *http.Client
	if strings.HasPrefix(endpoint, "unix://") {
		// Unix socket
		client = &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					var d net.Dialer
					return d.DialContext(ctx, "unix", endpoint)
				},
			},
			Timeout: timeout,
		}
	} else {
		// IP:Port
		client = &http.Client{
			Timeout: timeout,
		}
	}
	return &ClientFactory{
		client: client,
	}
}

func (f *ClientFactory) GetClient() *http.Client {
	return f.client
}

func SendEvent(factory *ClientFactory, event events.Status, message, endpoint string) {
	uri := fmt.Sprintf("http://%s/notify?event=%s&message=%s", endpoint, event, url.QueryEscape(message))
	client := factory.GetClient()
	logrus.Debugf("notify %s event to %s", event, uri)
	resp, err := client.Get(uri)
	if err != nil {
		logrus.Errorf("Error sending request: %v", err)
		return
	}
	defer resp.Body.Close()
}
