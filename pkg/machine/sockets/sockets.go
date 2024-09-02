package sockets

import (
	"bauklotze/pkg/fileutils"
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

// WaitForSocketWithBackoffs attempts to discover listening socket in maxBackoffs attempts
func WaitForSocketWithBackoffs(maxBackoffs int, backoff time.Duration, socketPath string, name string) error {
	backoffWait := backoff
	logrus.Debugf("checking that %q socket is ready", name)
	for i := 0; i < maxBackoffs; i++ {
		err := fileutils.Exists(socketPath)
		if err == nil {
			return nil
		}
		time.Sleep(backoffWait)
		backoffWait *= 2
	}
	return fmt.Errorf("unable to connect to %q socket at %q", name, socketPath)
}

// ListenAndWaitOnSocket waits for a new connection to the listener and sends
// any error back through the channel. ListenAndWaitOnSocket is intended to be
// used as a goroutine
func ListenAndWaitOnSocket(errChan chan<- error, listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		logrus.Debug("failed to connect to ready socket")
		errChan <- err
		return
	}
	_, err = bufio.NewReader(conn).ReadString('\n')
	logrus.Debug("ready ack received")

	if closeErr := conn.Close(); closeErr != nil {
		errChan <- closeErr
		return
	}

	errChan <- err
}
