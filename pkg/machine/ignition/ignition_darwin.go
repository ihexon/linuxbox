//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0
//go:build darwin

package ignition

import (
	"bauklotze/pkg/machine/defconfig"
	"bauklotze/pkg/machine/io"
	"bauklotze/pkg/machine/vmconfig"
	"fmt"
	"path/filepath"
)

func GenerateIgnScripts(mc *vmconfig.MachineConfig) error {
	var ignScriptFile = filepath.Join("/tmp", "initfs", "ovm_ign.sh")
	ign := NewIgnitionBuilder(
		&DynamicIgnitionV3{
			CodeBuffer: nil,
			IgnFile: io.FileWrapper{
				Path: ignScriptFile,
			},
			VMType: defconfig.LibKrun,
			Mounts: mc.Mounts,
			SSHIdentityPath: io.FileWrapper{
				Path: mc.SSH.IdentityPath,
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
