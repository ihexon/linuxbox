//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package service

import (
	"context"
	"fmt"

	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfig"
	"bauklotze/pkg/ssh"

	"github.com/sirupsen/logrus"
	sshSingal "golang.org/x/crypto/ssh"
)

func runCtx(ctx context.Context, mc *vmconfig.MachineConfig, name string, args []string) error {
	sshConfig, err := ssh.NewConfig(define.LocalHostURL, mc.SSH.RemoteUsername, uint(mc.SSH.Port), mc.SSH.IdentityPath)
	if err != nil {
		return fmt.Errorf("failed to create ssh config: %w", err)
	}
	myCmd := ssh.NewCmd(sshConfig)
	myCmd.SetCmdLine(ctx, name, args)

	myCmd.SetStopSignal(sshSingal.SIGKILL)

	logrus.Infof("SSH client running command: %s", myCmd.String())
	return myCmd.RunCtx() //nolint:wrapcheck
}

func run(mc *vmconfig.MachineConfig, name string, args []string) error {
	ctx := context.Background()
	return runCtx(ctx, mc, name, args)
}
