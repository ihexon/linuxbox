//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package machine

import (
	"crypto/sha256"
	"encoding/hex"

	"bauklotze/pkg/machine/vmconfigs"
)

type VolumeKind string

var (
	VirtIOFsVk VolumeKind = "virtiofs"
)

func (v VirtIoFs) ToMount() vmconfigs.Mount {
	return vmconfigs.Mount{
		ReadOnly: v.ReadOnly,
		Source:   v.Source,
		Tag:      v.Tag,
		Target:   v.Target,
		Type:     v.Kind(),
	}
}

func (v VirtIoFs) Kind() string {
	return string(VirtIOFsVk)
}

type VirtIoFs struct {
	VolumeKind
	ReadOnly bool
	Source   string
	Tag      string
	Target   string
}

// generateTag generates a tag for VirtIOFs mounts.
// AppleHV requires tags to be 36 bytes or fewer.
// SHA256 the path, then truncate to 36 bytes
func (v VirtIoFs) generateTag() string {
	sum := sha256.Sum256([]byte(v.Target))
	stringSum := hex.EncodeToString(sum[:])
	return stringSum[:36]
}

func NewVirtIoFsMount(src, target string, readOnly bool) VirtIoFs {
	vfs := VirtIoFs{
		ReadOnly: readOnly,
		Source:   src,
		Target:   target,
	}
	vfs.Tag = vfs.generateTag()
	return vfs
}
