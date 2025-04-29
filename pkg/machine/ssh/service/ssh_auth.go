//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package service

import (
	"context"
	"fmt"

	"bauklotze/pkg/machine/define"
	io2 "bauklotze/pkg/machine/io"

	"github.com/oomol-lab/ovm-ssh-agent/v3/pkg/identity"
	"github.com/oomol-lab/ovm-ssh-agent/v3/pkg/sshagent"
	system2 "github.com/oomol-lab/ovm-ssh-agent/v3/pkg/system"
	forwarder "github.com/oomol-lab/ssh-forward"
	"github.com/sirupsen/logrus"
)

type SSHAuthService struct {
	localSocks               string
	remoteSocks              string
	sshUser                  string
	sshAuthKey               string
	sshPort                  int
	errChanSSHAuth           chan error
	errChanUnixSocketForward chan error
}

// NewSSHAuthService a ssh agent forward service.
// listen a local socks file and forward to the upstream ssh agent socks.
func NewSSHAuthService(localSocks, remoteSocks, user, key string, port int) *SSHAuthService {
	return &SSHAuthService{
		localSocks:               localSocks,
		remoteSocks:              remoteSocks,
		sshUser:                  user,
		sshAuthKey:               key,
		sshPort:                  port,
		errChanSSHAuth:           make(chan error, 1),
		errChanUnixSocketForward: make(chan error, 1),
	}
}

func (s *SSHAuthService) StartSSHAuthServiceAndForwardV2(ctx context.Context) error {
	localSocketFile := io2.NewFile(s.localSocks)
	if err := localSocketFile.Delete(true); err != nil {
		return fmt.Errorf("failed to delete local ssh auth socks file: %w", err)
	}

	upstreamSocket := system2.GetSSHAgent()
	if upstreamSocket == "" {
		return fmt.Errorf("upstream SSH agent socket empty")
	}
	logrus.Infof("upstream ssh agent listened in: %q", upstreamSocket)

	ooSSHAgent := sshagent.NewSSHAgent(ctx, upstreamSocket, localSocketFile.GetPath())

	ooSSHAgent.LoadLocalKeys(identity.FindPrivateKeys()...)

	return ooSSHAgent.Serve() //nolint:wrapcheck
}

func (s *SSHAuthService) StartUnixSocketForward(ctx context.Context) error {
	logrus.Infof("forward unix socket %q to %q", s.localSocks, s.remoteSocks)
	socketForwarder := forwarder.NewUnixRemote(s.localSocks, define.LocalHostURL, s.remoteSocks).
		SetKeyFile(s.sshAuthKey).
		SetUser(s.sshUser).
		SetPort(s.sshPort)

	socketForwarder.SetConnState(func(tun *forwarder.ForwardConfig, state forwarder.ConnState) {
		switch state {
		case forwarder.StateStarting:
			logrus.Infoln("socket forwarder state is: starting")
		case forwarder.StateStarted:
			logrus.Infoln("socket forwarder state is: started")
		case forwarder.StateStopped:
			logrus.Infoln("socket forwarder state is: stopped")
		}
	})

	return socketForwarder.Start(ctx) //nolint:wrapcheck
}
