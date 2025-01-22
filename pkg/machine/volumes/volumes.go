//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package volumes

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type VolumeMountType int

const (
	VirtIOFS VolumeMountType = iota
)

func extractSourcePath(paths []string) string {
	return paths[0] + "/" // Add trailing slash to source path
}

func (v VolumeMountType) String() string {
	switch v {
	case VirtIOFS:
		return "virtiofs"
	default:
		return "unknown"
	}
}

func extractMountOptions(paths []string) bool {
	readonly := false
	if len(paths) > 2 { //nolint:mnd
		options := paths[2]
		volopts := strings.Split(options, ",")
		for _, o := range volopts {
			switch {
			case o == "rw":
				readonly = false
			case o == "ro":
				readonly = true
			default:
				fmt.Printf("Unknown option: %s\n", o)
			}
		}
	}
	return readonly
}

func SplitVolume(idx int, volume string) (string, string, string, bool) {
	tag := fmt.Sprintf("vol%d", idx)
	paths := pathsFromVolume(volume)
	source := extractSourcePath(paths)
	target := extractTargetPath(paths)
	readonly := extractMountOptions(paths)
	return tag, source, target, readonly
}

type Mount struct {
	ReadOnly bool   `json:"ReadOnly"`
	Source   string `json:"Source"`
	Tag      string `json:"Tag"`
	Target   string `json:"Target"`
	Type     string `json:"Type"`
}

func CmdLineVolumesToMounts(volumes []string) []*Mount {
	mounts := []*Mount{}
	for i, volume := range volumes {
		if volume == "" {
			continue
		}
		_, source, target, readOnly := SplitVolume(i, volume)
		m := NewVirtIoFsMount(source, target, readOnly).ToMount()
		mounts = append(mounts, &m)
	}
	return mounts
}

func (v VirtIoFs) ToMount() Mount {
	return Mount{
		ReadOnly: v.ReadOnly,
		Tag:      v.Tag,
		Source:   v.Source,
		Target:   v.Target,
		Type:     v.Kind(),
	}
}

const virtIOFsVk = "virtiofs"

func (v VirtIoFs) Kind() string {
	return virtIOFsVk
}

type VirtIoFs struct {
	ReadOnly bool
	Tag      string
	Source   string
	Target   string
}

// generateTag generates a tag for VirtIOFs mounts.
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
