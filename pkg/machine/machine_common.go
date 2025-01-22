//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package machine

import (
	"fmt"
	"time"

	"bauklotze/pkg/httpclient"

	"github.com/sirupsen/logrus"
)

const defaultPingTimeout = 5 * time.Second
const defaultPingInterval = 200 * time.Millisecond

func WaitAndPingAPI(sock string) error {
	client := httpclient.New().SetTransport(httpclient.CreateUnixTransport(sock))
	timeout := time.After(defaultPingTimeout)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout reached while waiting for Podman API")
		default:
			logrus.Info("Try ping Podman API")
			time.Sleep(defaultPingInterval)

			if err := client.Get("_ping"); err == nil {
				logrus.Infof("Podman ping test success")
				return nil
			}
		}
	}
}
