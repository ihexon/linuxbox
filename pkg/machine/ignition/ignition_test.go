//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package ignition

import (
	"bauklotze/pkg/machine/defconfig"
	"bauklotze/pkg/machine/io"
	"bauklotze/pkg/machine/volumes"
	"path/filepath"
	"testing"
)

func TestNewIgnitionBuilder(t *testing.T) {
	mycodes := []string{"echo END OF SCRIPT"}

	ign := NewIgnitionBuilder(
		&DynamicIgnitionV3{
			CodeBuffer: nil,
			IgnFile: io.FileWrapper{
				Path: filepath.Join("/tmp", "initfs", "ign.sh"),
			},
			VMType: defconfig.LibKrun,
			Mounts: []*volumes.Mount{
				{
					Type:   volumes.VirtIOFS.String(),
					Source: "/zzh",
					Tag:    "virtio-zzh",
					Target: "/mnt/zzh",
				},
				{
					Type:   volumes.VirtIOFS.String(),
					Source: "/zzh1",
					Tag:    "virtio-zzh1",
					Target: "/mnt/zzh1",
				},
				{
					Type:   volumes.VirtIOFS.String(),
					Source: "/zzh2",
					Tag:    "virtio-zzh2",
					Target: "/mnt/zzh2",
				},
			},
		})

	err := ign.GenerateIgnitionConfig(mycodes)
	if err != nil {
		t.Errorf("failed to generate ignition config: %v", err)
	} else {
		t.Log("Ignition config generated successfully")
		t.Log(ign.CodeBuffer.String())
	}

	err = ign.Write()
	if err != nil {
		t.Errorf("failed to write ignition file: %v", err)
	}
}
