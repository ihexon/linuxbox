//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build amd64 || arm64

package machine

import (
	"fmt"

	whatProvider "bauklotze/pkg/machine/provider"
	"bauklotze/pkg/machine/vmconfigs"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var provider vmconfigs.VMProvider

func machinePreRunE(cmd *cobra.Command, args []string) error {
	logrus.Infof("Try to get current hypervisor provider...")
	p, err := whatProvider.Get()
	if err != nil {
		return fmt.Errorf("failed to get current hypervisor provider: %w", err)
	}

	provider = p
	logrus.Infof("Got current hypervisor provider %s", p.VMType().String())
	return nil
}
