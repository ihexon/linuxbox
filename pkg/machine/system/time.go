//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package system

import (
	"fmt"
	"time"

	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/vmconfigs"
)

func TimeSync(mc *vmconfigs.MachineConfig) error {
	syncTimeCmd := fmt.Sprintf("date -s @%d", time.Now().Unix())
	if err := machine.CommonSSHSilent(mc.SSH.RemoteUsername, mc.SSH.IdentityPath, mc.Name, mc.SSH.Port, []string{syncTimeCmd}); err != nil {
		return fmt.Errorf("failed to sync timestamp: %w", err)
	}
	return nil
}
