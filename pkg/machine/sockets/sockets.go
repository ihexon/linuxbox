package sockets

import (
	"bauklotze/pkg/machine/ignition"
	"bauklotze/pkg/machine/vmconfigs"
	"bufio"
	"fmt"
	"github.com/containers/storage/pkg/fileutils"
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
		logrus.Errorf("failed to connect to ready socket")
		errChan <- err
		return
	}
	_, err = bufio.NewReader(conn).ReadString('\n')
	logrus.Infof("ready ack received")

	if closeErr := conn.Close(); closeErr != nil {
		errChan <- closeErr
		return
	}

	errChan <- err
}

func ListenAndExecCommandOnUnixSocketFile(listener net.Listener, mc *vmconfigs.MachineConfig) error {
	err := listener.(*net.UnixListener).SetDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return err
	}
	conn, err := listener.Accept()
	if err != nil {
		logrus.Errorf("the virtual machine failed to connect to the ignition socket in 10 sceonds, gvie up")
		return err
	}

	ignCfgs, err := ignition.ServeIgnitionOverSock(mc)
	if err != nil {
		return err
	}

	for _, c := range ignCfgs.Commands {
		logrus.Infof("ExecCommands: %s", c)
		_, err := conn.Write([]byte(c + "\n"))
		if err != nil {
			return err
		}
	}

	if err := conn.(*net.UnixConn).CloseWrite(); err != nil {
		return err
	}

	if closeErr := conn.Close(); closeErr != nil {
		return nil
	}

	return nil
}
