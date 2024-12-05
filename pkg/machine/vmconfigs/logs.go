//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package vmconfigs

import "bauklotze/pkg/machine/define"

func (mc *MachineConfig) LogFile() (*define.VMFile, error) {
	logsDir, err := mc.LogsDir()
	if err != nil {
		return nil, err
	}
	return logsDir.AppendToNewVMFile(mc.Name+".log", nil) //nolint:wrapcheck
}
