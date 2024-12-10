//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package vmconfigs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"bauklotze/pkg/machine/define"
)

func (mc *MachineConfig) GVProxySocket() (*define.VMFile, error) {
	machineRuntimeDir, err := mc.RuntimeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get runtime dir: %w", err)
	}
	return gvProxySocket(mc.Name, machineRuntimeDir)
}

func (mc *MachineConfig) PodmanAPISocketHost() (*define.VMFile, error) {
	machineRuntimeDir, err := mc.RuntimeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get runtime dir: %w", err)
	}
	return podmanAPISocketOnHost(mc.Name, machineRuntimeDir)
}

func (mc *MachineConfig) Lock() {
	mc.lock.Lock()
}

// Unlock removes an existing lock
func (mc *MachineConfig) Unlock() {
	mc.lock.Unlock()
}

// Refresh reloads the config file from disk
func (mc *MachineConfig) Refresh() error {
	content, err := os.ReadFile(mc.ConfigPath.GetPath())
	if err != nil {
		return fmt.Errorf("failed to read machine config: %w", err)
	}
	return json.Unmarshal(content, mc) //nolint:wrapcheck
}

// ConfigDir is a simple helper to obtain the machine config dir
func (mc *MachineConfig) ConfigDir() (*define.VMFile, error) {
	if mc.Dirs == nil || mc.Dirs.ConfigDir == nil {
		return nil, errors.New("no configuration directory set")
	}
	return mc.Dirs.ConfigDir, nil
}

func (mc *MachineConfig) UpdateLastBoot() error { //nolint:unused
	mc.LastUp = time.Now()
	return mc.Write()
}
