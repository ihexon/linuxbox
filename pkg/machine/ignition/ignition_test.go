//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package ignition

import (
	"path/filepath"
	"testing"

	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfigs"
)

func TestNewIgnitionBuilder(t *testing.T) {
	mycodes := []string{"echo END OF SCRIPT"}

	ign := NewIgnitionBuilder(
		&DynamicIgnitionV3{
			CodeBuffer: nil,
			IgnFile: define.VMFile{
				Path:    filepath.Join("/tmp", "initfs", "ign.sh"),
				Symlink: nil,
			},
			VMType: define.LibKrun,
			Mounts: []*vmconfigs.Mount{
				{
					Type:   vmconfigs.VirtIOFS.String(),
					Source: "/zzh",
					Tag:    "virtio-zzh",
					Target: "/mnt/zzh",
				},
				{
					Type:   vmconfigs.VirtIOFS.String(),
					Source: "/zzh1",
					Tag:    "virtio-zzh1",
					Target: "/mnt/zzh1",
				},
				{
					Type:   vmconfigs.VirtIOFS.String(),
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
