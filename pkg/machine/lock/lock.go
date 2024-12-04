//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package lock

import (
	"fmt"
	"path/filepath"

	"github.com/containers/storage/pkg/lockfile"
)

func GetMachineLock(name string, machineConfigDir string) (*lockfile.LockFile, error) {
	lockPath := filepath.Join(machineConfigDir, name+".lock")
	lock, err := lockfile.GetLockFile(lockPath)
	if err != nil {
		return nil, fmt.Errorf("creating lockfile for VM: %w", err)
	}
	return lock, nil
}
