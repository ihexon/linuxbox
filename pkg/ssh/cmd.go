//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type Cmd struct {
	// Path to command executable filename
	name string
	// Command args.
	args []string
	// SSH session.
	mySession *ssh.Session
	// Context for cancellation
	context context.Context
	// Signal send when the context is canceled
	signal ssh.Signal
	// ssh client configure
	config *Config
}

// SetStopSignal sets the signal to send when the context is canceled.
func (c *Cmd) SetStopSignal(signal ssh.Signal) {
	c.signal = signal
}

// SetCmdLine sets the command line to execute.
func (c *Cmd) SetCmdLine(ctx context.Context, name string, args []string) {
	c.name = name
	c.args = args
	c.context = ctx
}

const tcpProto = "tcp"

// NewClient New starts a new ssh connection, the host public key must be in known hosts.
func (c *Cmd) newClient() (*ssh.Client, error) {
	sshClient, err := ssh.Dial(
		tcpProto,
		net.JoinHostPort(c.config.Addr, fmt.Sprint(c.config.Port)),
		&ssh.ClientConfig{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			User:            c.config.User,
			Auth:            c.config.Auth,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ssh client:%w", err)
	}

	return sshClient, nil
}

// String returns the command line string, with each parameter wrapped in ""
func (c *Cmd) String() string {
	args := append([]string{c.name}, c.args...)
	for i, s := range args {
		args[i] = fmt.Sprintf("\"%s\"", s)
	}
	return strings.Join(args, " ")
}

// RunCtx executes the given callback within session. Sends SIGINT when the context is canceled.
func (c *Cmd) RunCtx() error {
	context.AfterFunc(c.context, func() {
		if c.mySession != nil {
			logrus.Warnf("send signal [ %s ] to [ %q ], cause by %v", c.signal, c.name, context.Cause(c.context))
			if err := c.mySession.Signal(c.signal); err != nil {
				logrus.Errorf("send signal [ %s ] to [ %q ] failed: %v", c.signal, c.name, err)
			}
		}
	})

	client, err := c.newClient()
	if err != nil {
		return fmt.Errorf("failed to create ssh client: %w", err)
	}
	defer client.Close() //nolint:errcheck

	c.mySession, err = client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create ssh session: %w", err)
	}

	outPipe, err := c.mySession.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get session.StdoutPipe(): %w", err)
	}
	errPipe, err := c.mySession.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get session.StderrPipe(): %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2) //nolint:mnd
	logStdErr := func(pipe io.Reader) {
		_, err := io.Copy(os.Stderr, pipe)
		if err != nil {
			logrus.Errorf("failed to copy pipe into os.Stderr")
		}
		wg.Done()
	}

	logStdOut := func(pipe io.Reader) {
		_, err := io.Copy(os.Stdout, pipe)
		if err != nil {
			logrus.Errorf("failed to copy pipe into os.Stdout")
		}
		wg.Done()
	}

	if err = c.mySession.Start(c.String()); err != nil {
		return fmt.Errorf("failed to start ssh command: %w", err)
	}

	defer func() {
		_ = c.mySession.Close()
		c.mySession = nil
	}()

	go logStdOut(outPipe)
	go logStdErr(errPipe)

	wg.Wait()

	if err = c.mySession.Wait(); err != nil {
		return fmt.Errorf("failed to wait ssh command: %w", err)
	}

	return nil
}
