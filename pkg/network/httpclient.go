package network

import (
	"bauklotze/pkg/events"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"time"
)

func NewUnixSocketClient(socketPath string, timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, "unix", socketPath)
			},
		},
		Timeout: timeout,
	}
}

func SendEvent(event events.Status, message, socks string) {
	client := NewUnixSocketClient(socks, 200*time.Millisecond)
	logrus.Debugf("notify %s event to %s", event, socks)
	resp, _ := client.Get(socks)
	if resp != nil {
		_ = resp.Body.Close()
	}
}

func SendEventToOvmJs(event events.Status, message string) {
	uri := fmt.Sprintf("http://ovm/notify?event=%s&message=%s", event, message)
	client := NewUnixSocketClient(uri, 200*time.Millisecond)
	logrus.Debugf("notify %s event to %s", event, uri)
	resp, _ := client.Get(uri)
	if resp != nil {
		_ = resp.Body.Close()
	}
}
