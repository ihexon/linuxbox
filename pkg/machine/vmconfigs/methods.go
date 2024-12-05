//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package vmconfigs

import (
	"encoding/json"
	"fmt"

	"bauklotze/pkg/machine/define"

	"github.com/containers/storage/pkg/ioutils"
)

// write is a non-locking way to write the machine configuration file to disk
func (mc *MachineConfig) Write() error {
	if mc.ConfigPath == nil {
		return fmt.Errorf("no configuration file associated with vm %q", mc.Name)
	}
	b, err := json.Marshal(mc)
	if err != nil {
		return fmt.Errorf("failed to marshal machine config: %w", err)
	}
	return ioutils.AtomicWriteFile(mc.ConfigPath.GetPath(), b, define.DefaultFilePerm) //nolint:wrapcheck
}
