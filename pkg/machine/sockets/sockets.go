//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package sockets

import (
	"fmt"
	"time"

	"github.com/containers/storage/pkg/fileutils"
	"github.com/sirupsen/logrus"
)

// WaitForSocketWithBackoffs attempts to discover listening socket in maxBackoffs attempts
func WaitForSocketWithBackoffs(socketPath string) error {
	var gvProxyWaitBackoff = 100 * time.Millisecond
	logrus.Infof("Checking that %s socket is ready\n", socketPath)
	for range 10 {
		err := fileutils.Exists(socketPath)
		if err == nil {
			return nil
		}
		logrus.Infof("Gvproxy Socket %s not ready, try again....\n", socketPath)
		time.Sleep(gvProxyWaitBackoff)
	}
	return fmt.Errorf("unable to connect to socket at %s", socketPath)
}
