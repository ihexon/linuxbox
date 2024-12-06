//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package sockets

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/containers/storage/pkg/fileutils"
	"github.com/sirupsen/logrus"
)

// WaitForSocketWithBackoffs attempts to discover listening socket in maxBackoffs attempts
func WaitForSocketWithBackoffs(socketPath string) error {
	var gvProxyWaitBackoff = 300 * time.Millisecond
	logrus.Infof("Checking that %s socket is ready\n", socketPath)
	for range 10 {
		err := fileutils.Exists(socketPath)
		if err == nil {
			return nil
		}
		logrus.Infof("Gvproxy Socket %s not ready, try again....\n", socketPath)
		time.Sleep(gvProxyWaitBackoff)
	}
	return fmt.Errorf("unable to connect to socket at %q", socketPath)
}

// ListenAndWaitOnSocket waits for a new connection to the listener and sends
// any error back through the channel. ListenAndWaitOnSocket is intended to be
// used as a goroutine
func ListenAndWaitOnSocket(errChan chan<- error, listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		logrus.Errorf("failed to connect to ready socket")
		errChan <- err
		return
	}
	_, err = bufio.NewReader(conn).ReadString('\n')
	logrus.Infof("READY ACK received")

	if closeErr := conn.Close(); closeErr != nil {
		errChan <- closeErr
		return
	}

	errChan <- err
}
