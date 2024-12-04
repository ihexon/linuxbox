//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package vmconfigs

import (
	"bauklotze/pkg/machine/define"
)

func (mc *MachineConfig) ReadySocket() (*define.VMFile, error) {
	rtDir, err := mc.RuntimeDir()
	if err != nil {
		return nil, err
	}
	return readySocket(mc.Name, rtDir)
}

func (mc *MachineConfig) IgnitionSocket() (*define.VMFile, error) {
	rtDir, err := mc.RuntimeDir()
	if err != nil {
		return nil, err
	}
	return ignitionSocket(mc.Name, rtDir)
}

func (mc *MachineConfig) CliProxyUDFAddr() (*define.VMFile, error) {
	return define.NewMachineFile("/tmp/cliproxy.sock", nil)
}
