//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0
//go:build darwin && arm64

package ignition

import (
	"fmt"
	"path/filepath"

	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfigs"
)

func GenerateIgnScripts(mc *vmconfigs.MachineConfig) error {
	var ignScriptFile = filepath.Join("/tmp", "initfs", "ovm_ign.sh")
	ign := NewIgnitionBuilder(
		&DynamicIgnitionV3{
			CodeBuffer: nil,
			IgnFile: define.VMFile{
				Path:    ignScriptFile,
				Symlink: nil,
			},
			VMType: define.LibKrun,
			Mounts: mc.Mounts,
			SSHIdentityPath: define.VMFile{
				Path:    mc.SSH.IdentityPath,
				Symlink: nil,
			},
		})

	err := ign.GenerateIgnitionConfig([]string{""})
	if err != nil {
		return fmt.Errorf("failed to generate ignition config: %w", err)
	}

	err = ign.Write()
	if err != nil {
		return fmt.Errorf("failed to write ignition file: %w", err)
	}

	return nil
}
