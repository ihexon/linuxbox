//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package shim

import (
	"bauklotze/pkg/machine/channel"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfig"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"

	"github.com/oomol-lab/ovm-ssh-agent/pkg/sshagent"
	"github.com/oomol-lab/ovm-ssh-agent/pkg/system"

	"github.com/oomol-lab/ovm-ssh-agent/pkg/identity"
	forwarder "github.com/oomol-lab/ssh-forward"

	"github.com/sirupsen/logrus"
)

func TryStartSSHAuthService(ctx context.Context, mc *vmconfig.MachineConfig) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("cancel TryStartSSHAuthService cause: %w", context.Cause(ctx))
	case <-channel.WaitVMReady():
		break
	}

	logrus.Infoln("Start SSH auth service")
	return startSSHAuthServiceAndForward(ctx, mc)
}

func startSSHAuthServiceAndForward(ctx context.Context, mc *vmconfig.MachineConfig) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(context.Canceled)

	g, ctx := errgroup.WithContext(ctx)

	localSocketFile := filepath.Join(mc.Dirs.SocksDir.GetPath(), "oo-ssh-agent-host.sock")
	if err := os.Remove(localSocketFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove local ssh agent socket: %w", err)
	}

	listener, err := net.Listen("unix", localSocketFile)
	if err != nil {
		return fmt.Errorf("failed to listen unix socket: %w", err)
	}
	defer listener.Close()

	upstreamSocket := system.GetSSHAgent()
	if upstreamSocket == "" {
		return fmt.Errorf("upstream SSH agent socket empty")
	}
	logrus.Infof("upstream ssh agent listened in: %q", upstreamSocket)

	ooSSHAgent, err := sshagent.NewSSHAgent(ctx, upstreamSocket)
	if err != nil {
		return fmt.Errorf("failed to create oo ssh agent: %w", err)
	}
	defer ooSSHAgent.Close()

	// find local private keys ~/.ssh
	ooSSHAgent.LoadLocalKeys(identity.FindPrivateKeys()...)

	g.Go(func() error {
		return ooSSHAgent.Serve(listener)
	})

	remoteSocketFile := "/opt/ssh_auth/oo-ssh-agent.sock"
	logrus.Infof("forward unix socket %q to %q", localSocketFile,
		fmt.Sprintf("%s@%s:%d:[%s]", mc.SSH.RemoteUsername, define.LocalHostURL, mc.SSH.Port, remoteSocketFile))

	socketForwarder := forwarder.NewUnixRemote(localSocketFile, define.LocalHostURL, remoteSocketFile)
	socketForwarder.SetTunneledConnState(func(tun *forwarder.ForwardConfig, state *forwarder.TunneledConnState) {
		logrus.Infof("connect state: %v", state)
	})

	socketForwarder.
		SetKeyFile(mc.SSH.IdentityPath).
		SetUser(mc.SSH.RemoteUsername).
		SetPort(mc.SSH.Port)

	// We set a callback to know when the tunnel is ready
	socketForwarder.SetConnState(func(tun *forwarder.ForwardConfig, state forwarder.ConnState) {
		switch state {
		case forwarder.StateStarting:
			logrus.Infof("socket forwarder state is staring")
			logrus.Infof("clean target socket file:%s", socketForwarder.Remote.String())
			if err := socketForwarder.CleanTargetSocketFile(); err != nil {
				cancel(fmt.Errorf("failed to clean target socket file: %w", err))
			}
		case forwarder.StateStarted:
			logrus.Infoln("socket forwarder state is: started")
		case forwarder.StateStopped:
			logrus.Infoln("socket forwarder state is: stopped")
		}
	})

	g.Go(func() error {
		return socketForwarder.Start(ctx)
	})

	return g.Wait() //nolint:wrapcheck
}
