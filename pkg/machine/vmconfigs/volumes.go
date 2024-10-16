package vmconfigs

import (
	"fmt"
	"strings"
)

type VolumeMountType int

const (
	// 9pfs
	VirtIOFS VolumeMountType = iota
	RawDisk
	Unknown
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

func extractMountOptions(paths []string) (bool, string) {
	readonly := false
	securityModel := "none"
	if len(paths) > 2 {
		options := paths[2]
		volopts := strings.Split(options, ",")
		for _, o := range volopts {
			switch {
			case o == "rw":
				readonly = false
			case o == "ro":
				readonly = true
			case strings.HasPrefix(o, "security_model="):
				securityModel = strings.Split(o, "=")[1]
			default:
				fmt.Printf("Unknown option: %s\n", o)
			}
		}
	}
	return readonly, securityModel
}

func SplitVolume(idx int, volume string) (string, string, string, bool, string) {
	tag := fmt.Sprintf("vol%d", idx)
	paths := pathsFromVolume(volume)
	source := extractSourcePath(paths)
	target := extractTargetPath(paths)
	readonly, securityModel := extractMountOptions(paths)
	return tag, source, target, readonly, securityModel
}
